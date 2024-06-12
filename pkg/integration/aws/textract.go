package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
)

type Textract struct {
	svc *textract.Client
}

func InitTextract(aws aws.Config) *Textract {
	svc := textract.NewFromConfig(aws)

	return &Textract{
		svc: svc,
	}
}

func (s *Textract) ExtractText(ctx context.Context, decodeFile []byte, bucketName, key *string) (*textract.DetectDocumentTextOutput, error) {
	doc := &types.Document{}

	if len(decodeFile) > 0 {
		doc.Bytes = decodeFile
	} else {
		doc.S3Object = &types.S3Object{
			Bucket: bucketName,
			Name:   key,
		}
	}

	output, err := s.svc.DetectDocumentText(ctx, &textract.DetectDocumentTextInput{
		Document: doc,
	})
	if err != nil {
		return nil, nil
	}

	return output, err
}

func (s *Textract) ExtractFormAndTable(ctx context.Context, decodeFile []byte, bucketName, key *string) (*textract.AnalyzeDocumentOutput, error) {
	doc := &types.Document{}

	if len(decodeFile) > 0 {
		doc.Bytes = decodeFile
	} else {
		doc.S3Object = &types.S3Object{
			Bucket: bucketName,
			Name:   key,
		}
	}
	output, err := s.svc.AnalyzeDocument(ctx, &textract.AnalyzeDocumentInput{
		Document:     doc,
		FeatureTypes: []types.FeatureType{types.FeatureTypeForms, types.FeatureTypeTables, types.FeatureTypeSignatures, types.FeatureTypeLayout},
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *Textract) ExtractID(ctx context.Context, decodeFile []byte, bucketName, key *string) (*textract.AnalyzeIDOutput, error) {
	doc := types.Document{}

	if len(decodeFile) > 0 {
		doc.Bytes = decodeFile
	} else {
		doc.S3Object = &types.S3Object{
			Bucket: bucketName,
			Name:   key,
		}
	}

	output, err := s.svc.AnalyzeID(ctx, &textract.AnalyzeIDInput{
		DocumentPages: []types.Document{
			doc,
		},
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}
