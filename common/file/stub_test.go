// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package file_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"htqrcode/common/file"
)

func TestStubStorage_StoreAndGet(t *testing.T) {
	stub := file.NewStub()

	want := []byte("hello world")
	url, err := stub.StoreFile(context.Background(), "statements/test.html", want)
	require.NoError(t, err)
	assert.Equal(t, "https://stub-storage.local/statements/test.html", url)

	got, ok := stub.GetFile("statements/test.html")
	require.True(t, ok, "stored file should be retrievable")
	assert.Equal(t, want, got)
}

func TestStubStorage_GetFile_NotFound(t *testing.T) {
	stub := file.NewStub()

	_, ok := stub.GetFile("missing.html")
	assert.False(t, ok, "missing file should report ok=false")
}

func TestStubStorage_StoreFile_DefensiveCopy(t *testing.T) {
	stub := file.NewStub()

	src := []byte("original")
	_, err := stub.StoreFile(context.Background(), "x.html", src)
	require.NoError(t, err)

	// Mutating the source after store must not affect stored content.
	src[0] = 'X'
	got, ok := stub.GetFile("x.html")
	require.True(t, ok)
	assert.Equal(t, []byte("original"), got, "stored content should be isolated from caller mutations")
}

func TestStubStorage_GetFile_DefensiveCopy(t *testing.T) {
	stub := file.NewStub()

	_, err := stub.StoreFile(context.Background(), "y.html", []byte("original"))
	require.NoError(t, err)

	// Mutating the slice returned by GetFile must not affect the stored content:
	// GetFile has to return a copy, not the slice it holds internally.
	got, ok := stub.GetFile("y.html")
	require.True(t, ok)
	got[0] = 'X'

	again, ok := stub.GetFile("y.html")
	require.True(t, ok)
	assert.Equal(t, []byte("original"), again, "GetFile should return a copy, so mutating its result leaves stored content intact")
}
