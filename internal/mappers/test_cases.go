package mappers

import (
	"rankode/internal/models"
	db "rankode/internal/repository"
)

func DbTestCasesToModelTestCase(cases []db.TaskTestCase) []models.TestCase {
	ret := make([]models.TestCase, len(cases))
	for i := range cases{
		ret[i].Id = cases[i].ID
		ret[i].InputFileName = cases[i].InputFile
		ret[i].OutputFileName = cases[i].OutputFile
		ret[i].Order = cases[i].CaseOrder
	}
	return ret
}