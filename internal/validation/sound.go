package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var allowedExtensions = map[string]struct{}{
	".wav": {},
	".mp3": {},
}

type AppError struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Details == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Details)
}

func NewAppError(message, details string) *AppError {
	return &AppError{
		Message: message,
		Details: details,
	}
}

type ResolvedSound struct {
	OriginalPath string
	ResolvedPath string
}

type SoundValidator struct {
	soundsDir string
}

func NewSoundValidator(workingDir string) *SoundValidator {
	return &SoundValidator{soundsDir: filepath.Join(workingDir, "sounds")}
}

func (v *SoundValidator) ValidateConfiguredPath(soundPath string) (ResolvedSound, *AppError) {
	return v.validatePath(
		soundPath,
		"configured soundPath must not be empty",
		"set the server startup argument --sound to a file name under the sounds directory",
	)
}

func (v *SoundValidator) ValidateRequestedPath(soundPath string) (ResolvedSound, *AppError) {
	return v.validatePath(
		soundPath,
		"soundPath must not be empty",
		"provide a tool argument soundPath or configure the server startup argument --sound",
	)
}

func (v *SoundValidator) ValidateOneShotPath(soundPath string) (ResolvedSound, *AppError) {
	return v.validatePath(
		soundPath,
		"play-once soundPath must not be empty",
		"set the command argument --play-once to a file name under the sounds directory",
	)
}

func (v *SoundValidator) validatePath(soundPath, emptyMessage, emptyDetails string) (ResolvedSound, *AppError) {
	if strings.TrimSpace(soundPath) == "" {
		return ResolvedSound{}, NewAppError(
			emptyMessage,
			emptyDetails,
		)
	}

	if filepath.IsAbs(soundPath) {
		return ResolvedSound{}, NewAppError(
			"absolute paths are not allowed",
			fmt.Sprintf("configure a file under the sounds directory: %s", v.soundsDir),
		)
	}

	resolvedPath := filepath.Clean(filepath.Join(v.soundsDir, soundPath))
	var err error
	resolvedPath, err = filepath.Abs(resolvedPath)
	if err != nil {
		return ResolvedSound{}, NewAppError(
			"failed to resolve configured soundPath",
			err.Error(),
		)
	}

	soundsDir, err := filepath.Abs(v.soundsDir)
	if err != nil {
		return ResolvedSound{}, NewAppError(
			"failed to resolve sounds directory",
			err.Error(),
		)
	}

	relativeToSounds, err := filepath.Rel(soundsDir, resolvedPath)
	if err != nil {
		return ResolvedSound{}, NewAppError(
			"failed to validate configured soundPath",
			err.Error(),
		)
	}

	if relativeToSounds == ".." || strings.HasPrefix(relativeToSounds, ".."+string(filepath.Separator)) {
		return ResolvedSound{}, NewAppError(
			"configured soundPath must stay within the sounds directory",
			fmt.Sprintf("allowed base directory: %s", soundsDir),
		)
	}

	extension := strings.ToLower(filepath.Ext(resolvedPath))
	if _, ok := allowedExtensions[extension]; !ok {
		return ResolvedSound{}, NewAppError(
			fmt.Sprintf("unsupported audio format: %s", extension),
			"currently supported audio formats are .wav and .mp3",
		)
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ResolvedSound{}, NewAppError(
				"configured sound file does not exist",
				fmt.Sprintf("looked under sounds directory: %s", resolvedPath),
			)
		}
		return ResolvedSound{}, NewAppError(
			"failed to inspect configured sound file",
			err.Error(),
		)
	}

	if !info.Mode().IsRegular() {
		return ResolvedSound{}, NewAppError(
			"configured soundPath must reference a regular file",
			fmt.Sprintf("resolved path: %s", resolvedPath),
		)
	}

	return ResolvedSound{
		OriginalPath: soundPath,
		ResolvedPath: resolvedPath,
	}, nil
}
