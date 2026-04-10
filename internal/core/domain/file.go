package domain

type File struct {
	buffer []byte
}

// Buffer возвращает байтовое содержимое файла
func (f *File) Buffer() []byte {
	return f.buffer
}

// NewFile создаёт новый File из переданного байтового буффера
func NewFile(buffer []byte) File {
	return File{
		buffer: buffer,
	}
}
