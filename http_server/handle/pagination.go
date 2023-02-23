package handle

type Pagination struct {
	Page    int `json:"page"`
	Size    int `json:"size"`
	maxSize int
}

func (p *Pagination) SetMaxSize(maxSize int) {
	p.maxSize = maxSize
}

func (p *Pagination) GetLimit() int {
	maxSize := p.maxSize
	if maxSize <= 0 {
		maxSize = 100
	}
	if p.Size < 1 || p.Size > maxSize {
		return maxSize
	}
	return p.Size
}

func (p *Pagination) GetOffset() int {
	page := p.Page
	if p.Page < 1 {
		page = 1
	}
	size := p.GetLimit()
	return (page - 1) * size
}
