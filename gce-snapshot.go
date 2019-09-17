package snapshot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

const maxNameLength = 63

// DryRun snapshotting
var DryRun bool

var computeService *compute.Service

func init() {
	client, err := google.DefaultClient(context.Background(), compute.ComputeScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err = compute.New(client)
	if err != nil {
		log.Fatal(err)
	}
}

// Create snapshot in GCE
func Create(ctx context.Context, projectName string) {
	disksService := compute.NewDisksService(computeService)
	req := disksService.AggregatedList(projectName)
	if err := req.Filter("labels.auto_snapshot:true").Pages(ctx, func(page *compute.DiskAggregatedList) error {
		for z, disks := range page.Items {
			for _, disk := range disks.Disks {
				if len(disk.Users) > 0 && disk.Status == "READY" {
					now := time.Now()
					snapshotName := fmt.Sprintf("%s-%d", disk.Name, now.UTC().Unix())
					if len(snapshotName) > maxNameLength {
						snapshotName = snapshotName[0:maxNameLength]
					}
					zone := strings.Split(z, "/")[1]
					snapshot := &compute.Snapshot{
						Description: fmt.Sprintf("Snapshot of %s at %s", disk.Name, now.Format(time.RFC822)),
						Name:        snapshotName,
						Labels:      disk.Labels,
					}
					log.Printf("[%s] Create snapshot %s", zone, snapshot.Name)
					if DryRun {
						log.Println("DRY-RUN")
					} else {
						_, err := disksService.CreateSnapshot(projectName, zone, disk.Name, snapshot).Do()
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("Disk list error: %s", err)
	}
}

// Purge snapshot in GCE
func Purge(ctx context.Context, projectName string, pruneDays, pruneMonth int) {
	snapshotsService := compute.NewSnapshotsService(computeService)
	req := snapshotsService.List(projectName)
	if err := req.Filter("labels.auto_snapshot:true").Pages(ctx, func(page *compute.SnapshotList) error {
		for _, snapshot := range page.Items {
			if t, err := time.Parse(time.RFC3339, snapshot.CreationTimestamp); err == nil {
				if t.Before(time.Now().AddDate(0, -pruneMonth, 0)) ||
					(t.Before(time.Now().AddDate(0, 0, -pruneDays)) && t.Day() != 1) {
					log.Printf("%s Purge snapshot %s created %s",
						snapshot.StorageLocations,
						snapshot.Name,
						snapshot.CreationTimestamp)
					if DryRun {
						log.Println("DRY-RUN")
					} else {
						if _, err := snapshotsService.Delete(projectName, snapshot.Name).Do(); err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("Snapshot list error: %s", err)
	}
}
