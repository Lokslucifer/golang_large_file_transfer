package services

import (
	"context"
	"large_fss/internals/constants"
	"log"
	"path/filepath"

	"github.com/robfig/cron/v3"
)

func (s *Service) CleanupService() {
	c := cron.New()

	// Run CleanFailedUploadsService every hour
	_, err := c.AddFunc("@every 1h", func() {
		if err := s.CleanFailedUploadsService(); err != nil {
			log.Printf("cron: error cleaning failed uploads: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("cron: failed to schedule CleanFailedUploadsService: %v", err)
	}

	// Run CleanExpiredTransfersService every hour
	_, err = c.AddFunc("@every 1h", func() {
		if err := s.CleanExpiredTransfersService(); err != nil {
			log.Printf("cron: error cleaning expired transfers: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("cron: failed to schedule CleanExpiredTransfersService: %v", err)
	}

	// Start the cron scheduler in the background
	c.Start()

	log.Println("cleanup service: cron jobs scheduled")

	// You might want to block here or integrate with your app's lifecycle
	// For example, if this is a standalone tool, block forever:
	select {}
}

func (s *Service) CleanFailedUploadsService() error {
	ctx := context.Background()
	failedtransfers, err := s.repo.FindAllFailedTempTransfers(context.Background())
	if err != nil {
		return err
	}
	for _, ftrans := range failedtransfers {
		transPath := filepath.Join(constants.ChunkDir, ftrans.ID.String())
		err := s.filestorage.DeleteAll(ctx, transPath)
		if err != nil {
			log.Printf("cleanfailed upload service: error in deleting fs of %s: %w", ftrans.ID, err)
			continue
		}
		err = s.repo.DeleteTempTransferByID(ctx, ftrans.ID)
		if err != nil {
			log.Printf("cleanfailed upload service: error in deleting db of %s: %w", ftrans.ID, err)

		}

	}
	return nil

}
func (s *Service) CleanExpiredTransfersService() error {

	ctx := context.Background()

	expiredtransfers, err := s.repo.FindAllExpiredTransfers(ctx)
	if err != nil {
		return err
	}
	for _, exptrans := range expiredtransfers {
		filelst, err := s.repo.FindAllFilesByTransferID(ctx, exptrans.ID)
		shouldSkip:=false
		for _, file := range filelst {
			if file.NumOfActiveStream > 0 {
				shouldSkip=true
				break
			}
		}
		if(shouldSkip){
			continue
		}
		err = s.filestorage.DeleteAll(ctx, exptrans.TransferPath)
		if err != nil {
			log.Printf("clean expired transfers service: error in deleting fs of %s: %w", exptrans.ID, err)
			continue
		}
		err = s.repo.DeleteTransferByID(ctx, exptrans.ID)
		if err != nil {
			log.Printf("clean expired transfers service: error in deleting db of %s: %w", exptrans.ID, err)

		}
	}
	return nil
}
