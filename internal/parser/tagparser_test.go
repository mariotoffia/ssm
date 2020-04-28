package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyPmsTagShallFail(t *testing.T) {
	_, err := ParsePmsTagString("", "", "dev", "test-service")

	assert.NotEqual(t, nil, err)
}

func TestEmptyAsmTagShallFail(t *testing.T) {
	_, err := ParseAsmTagString("", "", "dev", "test-service")

	assert.NotEqual(t, nil, err)
}

func TestSingleWordInAsmTagShallBeTranslatedToName(t *testing.T) {
	ssmTag, err := ParseAsmTagString("myproperty", "", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
}

func TestSingleWordInPmsTagShallBeTranslatedToName(t *testing.T) {
	ssmTag, err := ParsePmsTagString("myproperty", "", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
}

func TestSingleWordInPmsAndEnvIsDevAndServiceIsTestServiceTagShallRenderProperPrefixAndFullName(t *testing.T) {
	ssmTag, err := ParsePmsTagString("myproperty", "", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/test-service", ssmTag.Prefix())
	assert.Equal(t, "/dev/test-service/myproperty", ssmTag.FullName())
}

func TestSingleWordInAsmAndEnvIsDevAndServiceIsTestServiceTagShallRenderProperPrefixAndFullName(t *testing.T) {
	ssmTag, err := ParseAsmTagString("myproperty", "", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/test-service", ssmTag.Prefix())
	assert.Equal(t, "/dev/test-service/myproperty", ssmTag.FullName())
}

func TestAsmTagInparamPrefixIsAlwaysAfterEnvironmentAndService(t *testing.T) {
	ssmTag, err := ParseAsmTagString("myproperty", "/dummy-prefix", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/test-service/dummy-prefix/myproperty", ssmTag.FullName())
}

func TestPmsTagInparamPrefixIsAlwaysAfterEnvironmentAndService(t *testing.T) {
	ssmTag, err := ParsePmsTagString("myproperty", "/dummy-prefix", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/test-service/dummy-prefix/myproperty", ssmTag.FullName())
}

func TestPmsTagWithPrefixOverridesInparamPrefixAndServcice(t *testing.T) {
	ssmTag, err := ParsePmsTagString("myproperty, prefix=/baah/bii", "/dummy-prefix", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/baah/bii/myproperty", ssmTag.FullName())
}

func TestAsmTagWithPrefixOverridesInparamPrefixAndServcice(t *testing.T) {
	ssmTag, err := ParseAsmTagString("myproperty, prefix=/baah/bii", "/dummy-prefix", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/baah/bii/myproperty", ssmTag.FullName())
}

func TestPmsDoubleNonKvNamesWillFail(t *testing.T) {
	_, err := ParsePmsTagString("myproperty, hiihaa", "/dummy-prefix", "dev", "test-service")

	assert.NotEqual(t, nil, err)
}

func TestAsmDoubleNonKvNamesWillFail(t *testing.T) {
	_, err := ParseAsmTagString("myproperty, hiihaa", "/dummy-prefix", "dev", "test-service")

	assert.NotEqual(t, nil, err)
}

func TestAsmSingleNameAndKvNameShallFail(t *testing.T) {
	_, err := ParseAsmTagString("myproperty, name=hiihaaa", "/dummy-prefix", "dev", "test-service")

	assert.NotEqual(t, nil, err)
}

func TestPmsSingleNameAndKvNameShallFail(t *testing.T) {
	_, err := ParsePmsTagString("myproperty, name=hiihaaa", "/dummy-prefix", "dev", "test-service")

	assert.NotEqual(t, nil, err)
}

func TestPmsTagWithNameAndKeyIdShallBeSecure(t *testing.T) {
	ssmTag, err := ParsePmsTagString("myproperty, keyid=arn://12902309-121212-1299845", "", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/test-service/myproperty", ssmTag.FullName())
	assert.Equal(t, true, ssmTag.Secure())
}
func TestAsmExtraKvIsTags(t *testing.T) {
	ssmTag, err := ParseAsmTagString("myproperty, super=man,bibbo=blobban", "", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/test-service/myproperty", ssmTag.FullName())
	assert.Equal(t, 2, len(ssmTag.Tags()))
	assert.Equal(t, "man", ssmTag.Tags()["super"])
	assert.Equal(t, "blobban", ssmTag.Tags()["bibbo"])
}
func TestPmsExtraKvIsTags(t *testing.T) {
	ssmTag, err := ParsePmsTagString("myproperty, super=man,bibbo=blobban", "", "dev", "test-service")

	assert.Equal(t, nil, err)
	assert.Equal(t, "myproperty", ssmTag.Name())
	assert.Equal(t, "/dev/test-service/myproperty", ssmTag.FullName())
	assert.Equal(t, 2, len(ssmTag.Tags()))
	assert.Equal(t, "man", ssmTag.Tags()["super"])
	assert.Equal(t, "blobban", ssmTag.Tags()["bibbo"])
}
