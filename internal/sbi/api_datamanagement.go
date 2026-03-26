package sbi

import "github.com/gin-gonic/gin"

func (s *Server) HandleCreateDataStoreRecord(c *gin.Context) {
	s.Processor().HandleCreateDataStoreRecord(c)
}

func (s *Server) HandleCreateDataRetrievalSubscription(c *gin.Context) {
	s.Processor().HandleCreateDataRetrievalSubscription(c)
}

func (s *Server) HandleGetDataStoreRecords(c *gin.Context) {
	s.Processor().HandleGetDataStoreRecords(c)
}

func (s *Server) HandleDeleteDataRetrievalSubscription(c *gin.Context) {
	s.Processor().HandleDeleteDataRetrievalSubscription(c)
}
