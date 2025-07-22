package services

import (
	"context"

	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/models"
	"csort.ru/analysis-service/internal/repository"
)

var objectsServiceLog = logger.GetLogger("services.objects")

type ObjectsService struct {
	repo *repository.Queries
}

func NewObjectsService(repo *repository.Queries) *ObjectsService {
	return &ObjectsService{
		repo: repo,
	}
}

func (s *ObjectsService) GetObjects(ctx context.Context, objectIds []int32) ([]*models.ObjectMetadata, error) {
	rows, err := s.repo.GetObjectsMetadata(ctx, objectIds)
	if err != nil {
		return nil, err
	}
	objects := make([]*models.ObjectMetadata, 0, len(rows))
	for _, row := range rows {
		objects = append(objects, &models.ObjectMetadata{
			ID:       row.ID,
			MH:       row.MH.Float64,
			MS:       row.MS.Float64,
			MV:       row.MV.Float64,
			MR:       row.MR.Float64,
			MG:       row.MG.Float64,
			MB:       row.MB.Float64,
			LAvg:     row.LAvg.Float64,
			WAvg:     row.WAvg.Float64,
			BrtAvg:   row.BrtAvg.Float64,
			RAvg:     row.RAvg.Float64,
			GAvg:     row.GAvg.Float64,
			BAvg:     row.BAvg.Float64,
			HAvg:     row.HAvg.Float64,
			SAvg:     row.SAvg.Float64,
			VAvg:     row.VAvg.Float64,
			H:        row.H.Float64,
			S:        row.S.Float64,
			V:        row.V.Float64,
			HM:       row.HM.Float64,
			SM:       row.SM.Float64,
			VM:       row.VM.Float64,
			RM:       row.RM.Float64,
			GM:       row.GM.Float64,
			BM:       row.BM.Float64,
			BrtM:     row.BrtM.Float64,
			WM:       row.WM.Float64,
			LM:       row.LM.Float64,
			L:        row.L.Float64,
			W:        row.W.Float64,
			LW:       row.LW.Float64,
			Pr:       row.Pr.Float64,
			Sq:       row.Sq.Float64,
			Brt:      row.Brt.Float64,
			R:        row.R.Float64,
			G:        row.G.Float64,
			B:        row.B.Float64,
			Solid:    row.Solid.Float64,
			MinH:     row.MinH.Float64,
			MinS:     row.MinS.Float64,
			MinV:     row.MinV.Float64,
			MaxH:     row.MaxH.Float64,
			MaxS:     row.MaxS.Float64,
			MaxV:     row.MaxV.Float64,
			Entropy:  row.Entropy.Float64,
			ColorRhs: row.ColorRhs.String,
			SqSqcrl:  row.SqSqcrl.Float64,
			Hu1:      row.Hu1.Float64,
			Hu2:      row.Hu2.Float64,
			Hu3:      row.Hu3.Float64,
			Hu4:      row.Hu4.Float64,
			Hu5:      row.Hu5.Float64,
			Hu6:      row.Hu6.Float64,
		})
	}
	return objects, nil
}
