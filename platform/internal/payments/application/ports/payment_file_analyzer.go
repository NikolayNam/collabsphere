package ports

import (
	"context"
	"io"
)

// PaymentFileAnalyzer — анализатор платёжных файлов.
// Определяет формат, направление (входящие/исходящие) и парсит строки.
type PaymentFileAnalyzer interface {
	// DetectFormat пытается определить формат по имени файла и первым байтам.
	// Возвращает код формата (1c_bank, mt940, sberbank_csv, generic_csv) или пустую строку.
	DetectFormat(fileName string, peek []byte) string

	// Parse парсит содержимое файла и возвращает нормализованные строки.
	// formatCode — формат, если известен заранее; пустая строка — автоопределение.
	Parse(ctx context.Context, r io.Reader, fileName string, formatCode string) (*PaymentImportAnalysis, error)
}
