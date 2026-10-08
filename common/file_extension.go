package common

import (
	"errors"
	"image"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetFileExtension extracts the file extension from the given file path.
//
// The function takes a string representing the file path as input. It returns two values:
// - A string representing the file extension (without the dot).
// - An error, which is nil if the file path is valid, or an error message if the file path is invalid.
//
// The function uses the strings.LastIndex function to find the last occurrence of the dot in the file path.
// If the dot is not found, the function returns an empty string and an error message "invalid file path".
// Otherwise, it extracts the substring after the dot and returns it as the file extension.
func GetFileExtension(path string) (string, error) {
	lastIndex := strings.LastIndex(path, ".")
	if lastIndex == -1 {
		return "", errors.New("invalid file path")
	}
	return path[lastIndex+1:], nil
}

func GetFileExtension2(path string) (string, error) {
	ext := filepath.Ext(path)
	if ext == "" {
		return "", errors.New("No file extension")
	}
	return ext, nil
}

// GetTheRealFileExtension reads the file and returns its real file extension.
// It uses the http.DetectContentType function to determine the MIME type of the file.
// The function takes a multipart.File as input and returns the file extension as a string.
// If the file cannot be read or the MIME type cannot be determined, it returns an error message.
// The function uses the io.ReadAll function to read the entire file into memory.
// It then uses the mime.ExtensionsByType function to get the file extension based on the detected MIME type.
// If the MIME type is not recognized or the file cannot be read, it returns an error message.
// If return nil, nil => cannot detect extension. => Consider to use file extension from the original file name.
func GetTheRealFileExtension(file multipart.File) (string, error) {
	// Read the file and get the extension
	buffer, err := getTheFirst512ByteOfFile(file)
	if err != nil {
		return "", err
	}

	detectedExt := http.DetectContentType(buffer)
	ext, err := mime.ExtensionsByType(detectedExt)
	if err != nil {
		return "", errors.New("the file does not exist")
	}

	if len(ext) == 0 {
		return "", nil
	}
	return ext[0], nil
}

/**
* GetTheRealFileMimeType reads the file and returns its real MIME type.
* It uses the http.DetectContentType function to determine the MIME type of the file.
* The function takes a multipart.File as input and returns the MIME type as a string.
* If the file cannot be read or the MIME type cannot be determined, it returns an error message.
* The function reads the first 512 bytes of the file to determine its MIME type.
* If return nil, nil => cannot detect MIME type. => Consider to use file extension from the original file name.
* returns "image/jpeg" for JPEG files, "image/png" for PNG files, etc.
 */
func GetTheRealFileMimeType(file multipart.File) (string, error) {
	// Read the file and get the extension
	buffer, err := getTheFirst512ByteOfFile(file)
	if err != nil {
		return "", err
	}

	detectedExt := http.DetectContentType(buffer)
	return detectedExt, nil
}

func getTheFirst512ByteOfFile(file multipart.File) ([]byte, error) {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil || err == io.EOF {
		return nil, err
	}
	// Reset the reader to the beginning if need to re-read the file later.
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}
	return buffer, nil
}

// Get file MIME type from the file path.
// GetFileMimeType extracts the MIME type from the file path using the mime package.
// It returns the MIME type as a string and an error if the MIME type cannot be determined.
// return "application/octet-stream" if the MIME type cannot be determined. image/jpeg, image/png, video/mp4, etc.
func GetFileMimeType(file *os.File) (string, error) {
	// Read first 512 bytes (required by DetectContentType)
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Reset file offset to the beginning if needed
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// Detect the content type
	mimeType := http.DetectContentType(buffer[:n])
	return mimeType, nil
}

// Should not use this function to get file extension, because it is not safe to use incase you want to get file extension from the uploader.
// It's safe to get file extension from the validated file name or valid source.
// SplitFilePathAndFileExtension extracts the file path and file extension from the given file path.
func SplitFilePathAndFileExtension(path string) ([2]string, error) {
	lastIndex := strings.LastIndex(path, ".")
	if lastIndex == -1 {
		return [2]string{}, errors.New("invalid file path")
	}

	return [2]string{path[:lastIndex], path[lastIndex+1:]}, nil
	// return [2]string{path[:lastIndex], ext[0]}, nil
}

func SplitFilePathAndFileExtension2(path string) ([2]string, error) {
	fileExt := filepath.Ext(path)
	if fileExt == "" {
		return [2]string{}, errors.New("invalid file path")
	}
	fileNameTrimSuffix := strings.TrimSuffix(path, fileExt)
	if fileNameTrimSuffix == "" {
		return [2]string{}, errors.New("invalid file path")
	}

	return [2]string{fileNameTrimSuffix, fileExt}, nil
}

func GetImageResolution(file multipart.File) (int, int, error) {
	// Reset the reader to the beginning if need to re-read the file later.
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return 0, 0, err
		}
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return 0, 0, err
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Reset the reader to the beginning if need to re-read the file later.
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return 0, 0, err
		}
	}

	return width, height, nil
}
