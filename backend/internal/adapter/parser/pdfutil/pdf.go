package pdfutil

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func init() {
	// Disable pdfcpu's configuration directory creation to avoid errors
	// in restricted environments (like distroless/K8s) where $HOME/.config
	// cannot be created.
	api.DisableConfigDir()
}

// ExtractText pulls all readable text out of a PDF byte slice.
// It automatically attempts to decrypt the PDF using pdfcpu if it is encrypted
// (common for corporate payslips and bank statements with V=4/AES encryption).
func ExtractText(fileBytes []byte) (string, error) {
	// Try to decrypt PDF if it's encrypted (using pdfcpu)
	decryptedBytes, err := DecryptPDF(fileBytes)
	if err == nil {
		fileBytes = decryptedBytes
	}

	readerAt := bytes.NewReader(fileBytes)
	r, err := pdf.NewReader(readerAt, int64(len(fileBytes)))
	if err != nil {
		return "", fmt.Errorf("pdfutil: create reader: %w", err)
	}

	var buf strings.Builder
	totalPage := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue // skip unreadable pages
		}
		buf.WriteString(text)
		buf.WriteRune('\n')
	}
	return strings.TrimSpace(buf.String()), nil
}

// DecryptPDF attempts to remove PDF encryption using pdfcpu.
// It tries with an empty password, which is common for "protected" PDFs.
func DecryptPDF(fileBytes []byte) ([]byte, error) {
	conf := model.NewAESConfiguration("", "", 256)
	conf.ValidationMode = model.ValidationRelaxed

	var decrypted bytes.Buffer
	err := api.Decrypt(bytes.NewReader(fileBytes), &decrypted, conf)
	if err != nil {
		return nil, err
	}
	return decrypted.Bytes(), nil
}
