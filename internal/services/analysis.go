package services

import (
	"context"
	"fmt"
	"strconv"

	"bytes"
	"io"
	"mime/multipart"
	"net/http"

	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/models"
	"csort.ru/analysis-service/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
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
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("product", product)
	_ = writer.WriteField("userID", userID)

	fw, err := writer.CreateFormFile("files", fileName)
	if err != nil {
		return 0, nil, nil, err
	}
	if _, err := io.Copy(fw, fileContent); err != nil {
		return 0, nil, nil, err
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", s.analysisAPI, &buf)
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, resp.Header, nil, err
	}

	return resp.StatusCode, resp.Header, body, nil
}

func convertAnalysisFromRepo(repoAnalysis repository.Analysis) models.Analysis {
	idAnalysis, err := strconv.ParseInt(repoAnalysis.IDAnalysis.String, 10, 64)
	if err != nil {
		analysisLog.Error().Err(err).Str("idAnalysis", repoAnalysis.IDAnalysis.String).Msg("Failed to parse idAnalysis")
		return models.Analysis{}
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
		R:            repoAnalysis.R,
		G:            repoAnalysis.G,
		B:            repoAnalysis.B,
		H:            repoAnalysis.H,
		S:            repoAnalysis.S,
		V:            repoAnalysis.V,
		LabL:         repoAnalysis.LabL,
		LabA:         repoAnalysis.LabA,
		LabB:         repoAnalysis.LabB,
		W:            repoAnalysis.W,
		L:            repoAnalysis.L,
		T:            repoAnalysis.T,
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
