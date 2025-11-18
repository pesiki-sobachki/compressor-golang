package domain 

import "io"

type File struct{
	Content		io.ReadSeeker 	// Поток данных который можно перечитывать
	MimeType	string		//Тип контента, например "image/jpeg"
	Size		int64		//Размер в байтах
}

//Options определяет параметры для операции сжатия
type Options struct{
	Format 		string	//Целевой формат (webp, jpeg)
	Quality		int	//Качество сжатия
	MaxWidth	int	//Максимальная ширина
	MaxHeight	int	//Максимальная высота
}
