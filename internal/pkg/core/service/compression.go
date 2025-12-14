package service

import (
	"context"
	"fmt"
	"path/filepath" // Для работы с путями

	"github.com/andreychano/compressor-golang/internal/pkg/core/domain"
	"github.com/andreychano/compressor-golang/internal/pkg/core/port"
	"github.com/google/uuid" // для генерации уникальных имен
)

// CompressionService оркестрирует процесс компрессии.
type CompressionService struct {
	processors []port.Processor
	repository port.FileRepository
}

// NewCompressionService создает новый сервис. Мы передаем ему реализации портов.
func NewCompressionService(repo port.FileRepository, processors ...port.Processor) *CompressionService {
	return &CompressionService{
		repository: repo,
		processors: processors,
	}
}

/* CompressFile - основной метод, выполняющий всю работу. УСТАРЕЛ
func (s *CompressionService) CompressFile(ctx context.Context, file domain.File, opts domain.Options) (string, error) {
	var selectedProcessor port.Processor
	// 1. Найти подходящий процессор
	for _, p := range s.processors {
		if p.Supports(file.MimeType) {
			selectedProcessor = p
			break
		}
	}

	if selectedProcessor == nil {
		return "", fmt.Errorf("unsupported file type: %s", file.MimeType)
	}

	// 2. Обработать файл
	compressedFile, err := selectedProcessor.Process(file, opts)
	if err != nil {
		return "", fmt.Errorf("failed to process file: %w", err)
	}

	// 3. Сохранить файл
	// Генерируем уникальное имя, чтобы файлы не перезаписывали друг друга
	uniqueID := uuid.New().String()
	// Используем формат из опций для расширения файла
	fileName := fmt.Sprintf("%s.%s", uniqueID, opts.Format)
	filePath := filepath.Join("compressed", fileName)

	savedPath, err := s.repository.Save(ctx, compressedFile, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return savedPath, nil
}
*/

//Новые функции делят старую на части и еще добавлятся возврат пути GetFile

//1. Process - "Чистая функция"
//Только сжимает картинку и возвращает результат в памяти
//Ничего не сохраняет на диск

func (s *CompressionService) Process(file domain.File, opts domain.Options) (domain.File, error) {
	var selectedProcessor port.Processor

	//Подбор подходящего процессора компрессии

	for _, p := range s.processors {
		if p.Supports(file.MimeType) {
			selectedProcessor = p
			break
		}
	}

	if selectedProcessor == nil {
		return domain.File{}, fmt.Errorf("unsupport file type: %s", file.MimeType)
	}

	//Сжатие

	return selectedProcessor.Process(file, opts)
}

//2. Compress and save - "Полный цикл"
//Вызывает Process, потом отдает результат в Repository для последующего сохранения.

func (s *CompressionService) CompressAndSave(ctx context.Context, file domain.File, opts domain.Options) (string, error) {
	//Сжатие, использование метода Process
	compressedFile, err := s.Process(file, opts)
	if err != nil {
		return "", err
	}

	//Генерируем уникальное имя
	uniqueID := uuid.New().String()
	fileName := fmt.Sprintf("%s.%s", uniqueID, opts.Format)

	//Кладем в папку "compressed" внутри хранилища
	filePath := filepath.Join("compressed", fileName)

	//Сохраняем
	return s.repository.Save(ctx, compressedFile, filePath)
}

// 3. GetFile - "Поиск"
// Просит в repository найти файл
func (s *CompressionService) GetFile(ctx context.Context, path string) (domain.File, error) {
	return s.repository.Get(ctx, path)
}
