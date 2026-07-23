package sbi

import "github.com/gin-gonic/gin"

func (s *Server) HandleCreateMLModelStoreRecord(c *gin.Context) {
	s.Processor().HandleCreateMLModelStoreRecord(c)
}

func (s *Server) HandleGetMLModelStoreRecords(c *gin.Context) {
	s.Processor().HandleGetMLModelStoreRecords(c)
}

func (s *Server) HandleGetIndividualMLModelStoreRecord(c *gin.Context) {
	s.Processor().HandleGetIndividualMLModelStoreRecord(c)
}

func (s *Server) HandleDownloadMLModelFile(c *gin.Context) {
	s.Processor().HandleDownloadMLModelFile(c)
}

func (s *Server) HandleUpdateIndividualMLModelStoreRecord(c *gin.Context) {
	s.Processor().HandleUpdateIndividualMLModelStoreRecord(c)
}

func (s *Server) HandleDeleteIndividualMLModelStoreRecord(c *gin.Context) {
	s.Processor().HandleDeleteIndividualMLModelStoreRecord(c)
}
