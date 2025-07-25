package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"bytes"
	"io"
	"mime/multipart"
	"net/http"

	"encoding/json"

	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/models"
	"csort.ru/analysis-service/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/valyala/fasthttp"
)

const (
	DefaultLimit     = 10
	MaxLimit         = 100
	DefaultSortBy    = "date_time"
	DefaultSortOrder = "desc"
)

var analysisLog = logger.GetLogger("services.analysis")

type AnalysisService struct {
	repo        *repository.Queries
	analysisAPI string
}

func NewAnalysisService(repo *repository.Queries, analysisAPI string) *AnalysisService {
	return &AnalysisService{
		repo:        repo,
		analysisAPI: analysisAPI,
	}
}

func (s *AnalysisService) GetAnalyses(ctx context.Context, userID int64, params models.GetAnalysesPaginatedRequest) (*models.PaginatedResponse[models.Analysis], error) {
	// Set defaults
	if params.Limit == 0 {
		params.Limit = DefaultLimit
	}
	if params.Limit > MaxLimit {
		params.Limit = MaxLimit
	}
	if params.SortBy == "" {
		params.SortBy = DefaultSortBy
	}
	if params.SortOrder == "" {
		params.SortOrder = DefaultSortOrder
	}

	// Get analyses from repository
	repoAnalyses, err := s.repo.GetAnalysesByUserTelegramIDPagination(ctx, repository.GetAnalysesByUserTelegramIDPaginationParams{
		Limit:      params.Limit,
		Offset:     params.Offset,
		IDUser:     pgtype.Text{String: fmt.Sprintf("%d", userID), Valid: true},
		Product:    params.Product,
		IDAnalysis: params.ID,
		SortBy:     params.SortBy,
		SortOrder:  params.SortOrder,
	})
	if err != nil {
		analysisLog.Error().Err(err).Int64("userID", userID).Msg("Failed to get analyses")
		return nil, err
	}

	// Get total count
	count, err := s.repo.CountAnalysesByUserID(ctx, repository.CountAnalysesByUserIDParams{
		IDUser:     pgtype.Text{String: fmt.Sprintf("%d", userID), Valid: true},
		Product:    params.Product,
		IDAnalysis: params.ID,
	})
	if err != nil {
		analysisLog.Error().Err(err).Int64("userID", userID).Msg("Failed to count analyses")
		return nil, err
	}

	// Convert to service models
	analyses := make([]models.Analysis, 0, len(repoAnalyses))
	for _, repoAnalysis := range repoAnalyses {
		analyses = append(analyses, convertAnalysisFromRepo(repoAnalysis))
	}

	return &models.PaginatedResponse[models.Analysis]{
		Data:   analyses,
		Total:  count,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (s *AnalysisService) GetAnalysisByID(ctx context.Context, analysisID string) (models.Analysis, error) {
	// Get analysis
	repoAnalysis, err := s.repo.GetAnalysisByID(ctx, pgtype.Text{String: analysisID, Valid: true})
	if err != nil {
		analysisLog.Error().Err(err).Str("analysisID", analysisID).Msg("Failed to get analysis")
		return models.Analysis{}, err
	}

	// Get objects
	objects, err := s.getObjectsForAnalysis(ctx, int64(repoAnalysis.ID))
	if err != nil {
		return models.Analysis{}, err
	}

	// Convert and attach objects
	analysis := convertAnalysisFromRepo(repoAnalysis)
	analysis.Objects = objects

	return analysis, nil
}

func (s *AnalysisService) GetObjectsByAnalysisID(ctx context.Context, analysisID string) ([]models.Object, error) {
	internalID, err := strconv.ParseInt(analysisID, 10, 64)
	if err != nil {
		return nil, err
	}

	return s.getObjectsForAnalysis(ctx, internalID)
}

func (s *AnalysisService) getObjectsForAnalysis(ctx context.Context, analysisID int64) ([]models.Object, error) {
	repoObjects, err := s.repo.GetObjectsByAnalysisID(ctx, pgtype.Int8{Int64: analysisID, Valid: true})
	if err != nil {
		analysisLog.Warn().Err(err).Int64("analysisID", analysisID).Msg("Failed to get objects")
		return nil, err
	}

	if len(repoObjects) == 0 {
		return []models.Object{}, nil
	}

	objects := make([]models.Object, 0, len(repoObjects))
	for _, repoObject := range repoObjects {
		objects = append(objects, convertObjectFromRepo(repoObject))
	}

	return objects, nil
}

func (s *AnalysisService) ProxyAnalysisAPICall(ctx context.Context, product, userID, fileName string, fileContent io.Reader) (int, http.Header, []byte, error) {
	// Create a multipart form buffer
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add text fields
	if err := writer.WriteField("product", product); err != nil {
		return 0, nil, nil, err
	}

	if err := writer.WriteField("userID", userID); err != nil {
		return 0, nil, nil, err
	}

	// Create form file field
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return 0, nil, nil, err
	}

	// Copy file content to form
	_, err = io.Copy(part, fileContent)
	if err != nil {
		return 0, nil, nil, err
	}

	// Close the multipart writer
	writer.Close()

	// Create fasthttp request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(s.analysisAPI)
	req.Header.SetMethod("POST")
	req.Header.SetContentType(writer.FormDataContentType())
	req.SetBody(body.Bytes())

	// Create response object
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request with context support
	client := &fasthttp.Client{}
	err = client.DoTimeout(req, resp, 2*time.Minute)
	if err != nil {
		return 0, nil, nil, err
	}

	// Convert fasthttp response headers to http.Header
	headers := make(http.Header)
	resp.Header.VisitAll(func(key, value []byte) {
		headers.Add(string(key), string(value))
	})

	// Copy response body to avoid issues after response is released
	responseBody := make([]byte, len(resp.Body()))
	copy(responseBody, resp.Body())

	return resp.StatusCode(), headers, responseBody, nil
}

func convertAnalysisFromRepo(repoAnalysis repository.Analysis) models.Analysis {
	idAnalysis, err := strconv.ParseInt(repoAnalysis.IDAnalysis.String, 10, 64)
	if err != nil {
		analysisLog.Error().Err(err).Str("idAnalysis", repoAnalysis.IDAnalysis.String).Msg("Failed to parse idAnalysis")
		return models.Analysis{}
	}

	unmarshalStats := func(data []byte, fieldName string) models.Stats {
		var stats models.Stats
		if err := json.Unmarshal(data, &stats); err != nil {
			analysisLog.Error().Err(err).Str("field", fieldName).Msg("Failed to unmarshal stats")
			return models.Stats{}
		}
		return stats
	}

	return models.Analysis{
		ID:           repoAnalysis.ID,
		DateTime:     repoAnalysis.DateTime.Time,
		Product:      repoAnalysis.Product.String,
		ColorRhs:     repoAnalysis.ColorRhs.String,
		IDUser:       repoAnalysis.IDUser.String,
		TelegramLink: repoAnalysis.TelegramLink.String,
		Text:         repoAnalysis.Text.String,
		FileSource:   repoAnalysis.FileSource.String,
		ScaleMmPixel: repoAnalysis.ScaleMmPixel.Float64,
		Mass:         repoAnalysis.Mass.Float64,
		Area:         repoAnalysis.Area.Float64,
		R:            unmarshalStats(repoAnalysis.R, "R"),
		G:            unmarshalStats(repoAnalysis.G, "G"),
		B:            unmarshalStats(repoAnalysis.B, "B"),
		H:            unmarshalStats(repoAnalysis.H, "H"),
		S:            unmarshalStats(repoAnalysis.S, "S"),
		V:            unmarshalStats(repoAnalysis.V, "V"),
		LabL:         unmarshalStats(repoAnalysis.LabL, "LabL"),
		LabA:         unmarshalStats(repoAnalysis.LabA, "LabA"),
		LabB:         unmarshalStats(repoAnalysis.LabB, "LabB"),
		W:            unmarshalStats(repoAnalysis.W, "W"),
		L:            unmarshalStats(repoAnalysis.L, "L"),
		T:            unmarshalStats(repoAnalysis.T, "T"),
		FileOutput:   repoAnalysis.FileOutput.String,
		IDAnalysis:   idAnalysis,
	}
}

func convertObjectFromRepo(repoObject repository.Object) models.Object {
	return models.Object{
		ID:         repoObject.ID,
		IdAnalysis: repoObject.IDAnalysis.Int64,
		File:       repoObject.File.String,
		Class:      repoObject.Class.String,
		Geometry:   repoObject.Geometry.String,
		MH:         repoObject.MH.Float64,
		MS:         repoObject.MS.Float64,
		MV:         repoObject.MV.Float64,
		MR:         repoObject.MR.Float64,
		MG:         repoObject.MG.Float64,
		MB:         repoObject.MB.Float64,
		LAvg:       repoObject.LAvg.Float64,
		WAvg:       repoObject.WAvg.Float64,
		BrtAvg:     repoObject.BrtAvg.Float64,
		RAvg:       repoObject.RAvg.Float64,
		GAvg:       repoObject.GAvg.Float64,
		BAvg:       repoObject.BAvg.Float64,
		HAvg:       repoObject.HAvg.Float64,
		SAvg:       repoObject.SAvg.Float64,
		VAvg:       repoObject.VAvg.Float64,
		H:          repoObject.H.Float64,
		S:          repoObject.S.Float64,
		V:          repoObject.V.Float64,
		HM:         repoObject.HM.Float64,
		SM:         repoObject.SM.Float64,
		VM:         repoObject.VM.Float64,
		RM:         repoObject.RM.Float64,
		GM:         repoObject.GM.Float64,
		BM:         repoObject.BM.Float64,
		BrtM:       repoObject.BrtM.Float64,
		WM:         repoObject.WM.Float64,
		LM:         repoObject.LM.Float64,
		L:          repoObject.L.Float64,
		W:          repoObject.W.Float64,
		LW:         repoObject.LW.Float64,
		Pr:         repoObject.Pr.Float64,
		Sq:         repoObject.Sq.Float64,
		Brt:        repoObject.Brt.Float64,
		R:          repoObject.R.Float64,
		G:          repoObject.G.Float64,
		B:          repoObject.B.Float64,
		Solid:      repoObject.Solid.Float64,
		MinH:       repoObject.MinH.Float64,
		MinS:       repoObject.MinS.Float64,
		MinV:       repoObject.MinV.Float64,
		MaxH:       repoObject.MaxH.Float64,
		MaxS:       repoObject.MaxS.Float64,
		MaxV:       repoObject.MaxV.Float64,
		Entropy:    repoObject.Entropy.Float64,
		IDImage:    repoObject.IDImage.Int64,
		ColorRhs:   repoObject.ColorRhs.String,
		SqSqcrl:    repoObject.SqSqcrl.Float64,
		Hu1:        repoObject.Hu1.Float64,
		Hu2:        repoObject.Hu2.Float64,
		Hu3:        repoObject.Hu3.Float64,
		Hu4:        repoObject.Hu4.Float64,
		Hu5:        repoObject.Hu5.Float64,
		Hu6:        repoObject.Hu6.Float64,
	}
}
