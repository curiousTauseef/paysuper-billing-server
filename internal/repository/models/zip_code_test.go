package models

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	fuzz "github.com/google/gofuzz"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
	"time"
)

func Test_ZipCode_MapZipCodeToMgo_Ok(t *testing.T) {
	f := fuzz.New().SkipFieldsWithPattern(regexp.MustCompile("^(.*)At$"))
	original := &billingpb.ZipCode{}
	f.Fuzz(original)

	transformed, err := MapZipCodeToMgo(original)
	assert.NoError(t, err)

	converted, err := MapMgoToZipCode(transformed)
	assert.NoError(t, err)

	assert.ObjectsAreEqualValues(original, converted)
}

func Test_ZipCode_MapZipCodeToMgo_Error_CreatedAt(t *testing.T) {
	original := &billingpb.ZipCode{
		CreatedAt: &timestamp.Timestamp{Seconds: -1, Nanos: -1},
	}
	_, err := MapZipCodeToMgo(original)
	assert.Error(t, err)
}

func Test_ZipCode_MapMgoToZipCode_Ok(t *testing.T) {
	f := fuzz.New()
	original := &MgoZipCode{}
	f.Fuzz(original)

	transformed, err := MapMgoToZipCode(original)
	assert.NoError(t, err)

	converted, err := MapZipCodeToMgo(transformed)
	assert.NoError(t, err)

	assert.ObjectsAreEqualValues(original, converted)
}

func Test_ZipCode_MapMgoToZipCode_Error_CreatedAt(t *testing.T) {
	original := &MgoZipCode{
		CreatedAt: time.Time{}.AddDate(-10000, 0, 0),
	}
	_, err := MapMgoToZipCode(original)
	assert.Error(t, err)
}
