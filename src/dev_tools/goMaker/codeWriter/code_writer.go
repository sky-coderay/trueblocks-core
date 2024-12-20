package codeWriter

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
)

func WriteCode(existingFn, newCode string) (bool, error) {
	if len(strings.Trim(newCode, "\n\t\r")) == 0 {
		return false, nil
	}

	if !file.FileExists(existingFn) {
		if !strings.Contains(existingFn, "/generated/") {
			if !file.FolderExists(filepath.Dir(existingFn)) {
				logger.Fatal("Folder does not exist for file", existingFn)
			}
			logger.Info(colors.Yellow+"Creating", existingFn, strings.Repeat(" ", 20)+colors.Off)
		}
		return updateFile(existingFn, newCode)
	}

	tempFn := existingFn + ".new"
	defer func() {
		// delete the temp file when we're done
		os.Remove(tempFn)
	}()

	// save the new code in a temporary file
	file.StringToAsciiFile(tempFn, newCode)

	// extract the EXISTING_CODE from the existing file
	existingParts, err := extractExistingCode(existingFn)
	if err != nil {
		return false, fmt.Errorf("error extracting existing code: %v", err)
	}

	// apply the EXISTING_CODE to the new code, format the new code and
	// write it back to the original file (potentially destroying it)
	wasModified, err := applyTemplate(tempFn, existingParts)
	if err != nil {
		return false, fmt.Errorf("error applying template: %v %s", err, existingFn)
	}

	msg := LogMessage{
		MessageType: "Progress",
		Message:     existingFn,
	}
	if !strings.Contains(existingFn, "/generated/") && wasModified {
		msg.MessageType = "Info"
		msg.Message = fmt.Sprintf("Wrote %s", existingFn)
	}
	logChannel <- msg

	return wasModified, nil
}

func extractExistingCode(fileName string) (map[int]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	existingCode := make(map[int]string)
	scanner := bufio.NewScanner(file)

	codeSection := 0
	inExistingCode := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "// EXISTING_CODE") {
			if inExistingCode {
				existingCode[codeSection] += line + "\n"
				codeSection++
				inExistingCode = false
			} else {
				existingCode[codeSection] = line + "\n"
				inExistingCode = true
			}
		} else if inExistingCode {
			existingCode[codeSection] += line + "\n"
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return existingCode, nil
}

func applyTemplate(tempFn string, existingCode map[int]string) (bool, error) {
	defer os.Remove(tempFn) // we always try to remove this file

	ff, err := os.Open(tempFn)
	if err != nil {
		return false, err
	}

	isOpen := false
	lineCnt := 0
	codeSection := 0
	var buffer bytes.Buffer
	scanner := bufio.NewScanner(ff)
	for scanner.Scan() {
		line := scanner.Text()
		lineCnt++
		if strings.Contains(line, "// EXISTING_CODE") {
			if isOpen {
				isOpen = false
				codeSection++
			} else {
				if code, ok := existingCode[codeSection]; ok {
					buffer.WriteString(code)
					isOpen = true
				} else {
					return false, fmt.Errorf("missing // EXISTING_CODE section %d line %d", codeSection, lineCnt)
				}
			}
		} else {
			buffer.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	ff.Close()
	return updateFile(tempFn, buffer.String())
}

func updateFile(tempFn, newCode string) (bool, error) {
	formatted := newCode
	fileExt := strings.TrimPrefix(filepath.Ext(strings.TrimSuffix(tempFn, ".new")), ".")

	if fileExt == "go" {
		formattedBytes, err := format.Source([]byte(newCode))
		if err != nil {
			return showErroredCode(newCode, err)
		}
		formatted = string(formattedBytes)
	} else if hasPrettier() {
		var parser string
		switch fileExt {
		case "md":
			parser = "markdown"
		case "yaml", "yml":
			parser = "yaml"
		case "js", "jsx":
			parser = "babel"
		case "ts", "tsx":
			parser = "typescript"
		default:
			// do nothing
		}
		if parser != "" {
			utils.System("prettier -w --parser " + parser + " " + tempFn + " >/dev/null")
			formatted = file.AsciiFileToString(tempFn)
		}
	} else {
		logger.Warn("Prettier not found, skipping formatting for", tempFn, ". Install Prettier with `npm install -g prettier`.")
	}

	origFn := strings.Replace(tempFn, ".new", "", 1)
	if string(formatted) == file.AsciiFileToString(origFn) {
		return false, nil
	} else {
		file.StringToAsciiFile(origFn, string(formatted)) // modified code is in the original file
		return true, nil
	}
}

type LogMessage struct {
	MessageType string
	Message     string
}

var logChannel = make(chan LogMessage, 100) // Buffered channel

// syncLogger goroutine
func syncLogger() {
	rep := strings.Repeat(" ", 30)
	for logMsg := range logChannel {
		switch logMsg.MessageType {
		case "Progress":
			logger.Progress(true, colors.Green+logMsg.Message+colors.Off+rep)
		case "Info":
			logger.Info(colors.Yellow + logMsg.Message + colors.Off + rep)
		default:
			logger.Error("Unknown message type:", logMsg.MessageType)
		}
	}
}

func init() {
	go syncLogger()
}

func hasPrettier() bool {
	utils.System("which prettier >./found 2>/dev/null")
	defer os.Remove("./found")
	return file.FileExists("./found")
}

func showErroredCode(newCode string, err error) (bool, error) {
	logger.Error("Error formatting code:", colors.Red, err, colors.Off)
	logger.Info("Code that caused the error:")
	for i, line := range strings.Split(newCode, "\n") {
		logger.Info(fmt.Sprintf("%s%4d%s: %s", colors.Yellow, i+1, colors.Off, line))
	}
	return false, err
}
