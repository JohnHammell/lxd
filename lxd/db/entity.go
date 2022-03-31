//go:build linux && cgo && !agent
// +build linux,cgo,!agent

package db

import (
	"fmt"
	"strings"

	"github.com/lxc/lxd/lxd/db/cluster"
)

// ErrUnknownEntityID describes the unknown entity ID error
var ErrUnknownEntityID = fmt.Errorf("Unknown entity ID")

// GetURIFromEntity returns the URI for the given entity type and entity ID.
func (c *Cluster) GetURIFromEntity(entityType int, entityID int) (string, error) {
	if entityID == -1 || entityType == -1 {
		return "", nil
	}

	_, ok := cluster.EntityNames[entityType]
	if !ok {
		return "", fmt.Errorf("Unknown entity type")
	}

	var err error
	var uri string

	switch entityType {
	case cluster.TypeImage:
		var images []Image

		err = c.transaction(func(tx *ClusterTx) error {
			images, err = tx.GetImages(ImageFilter{})
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get images: %w", err)
		}

		for _, image := range images {
			if image.ID != entityID {
				continue
			}

			uri = fmt.Sprintf(cluster.EntityURIs[entityType], image.Fingerprint, image.Project)
			break
		}

		if uri == "" {
			return "", ErrUnknownEntityID
		}
	case cluster.TypeProfile:
		var profiles []Profile

		err = c.transaction(func(tx *ClusterTx) error {
			profiles, err = tx.GetProfiles(ProfileFilter{})
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get profiles: %w", err)
		}

		for _, profile := range profiles {
			if profile.ID != entityID {
				continue
			}

			uri = fmt.Sprintf(cluster.EntityURIs[entityType], profile.Name, profile.Project)
			break
		}

		if uri == "" {
			return "", ErrUnknownEntityID
		}
	case cluster.TypeProject:
		projects := make(map[int64]string)

		err = c.transaction(func(tx *ClusterTx) error {
			projects, err = tx.GetProjectIDsToNames()
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get project names and IDs: %w", err)
		}

		name, ok := projects[int64(entityID)]
		if !ok {
			return "", ErrUnknownEntityID
		}

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], name)
	case cluster.TypeCertificate:
		var certificates []Certificate

		err = c.transaction(func(tx *ClusterTx) error {
			certificates, err = tx.GetCertificates(CertificateFilter{})
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get certificates: %w", err)
		}

		for _, cert := range certificates {
			if cert.ID != entityID {
				continue
			}

			uri = fmt.Sprintf(cluster.EntityURIs[entityType], cert.Name)
			break
		}

		if uri == "" {
			return "", ErrUnknownEntityID
		}
	case cluster.TypeContainer:
		fallthrough
	case cluster.TypeInstance:
		var instances []Instance

		err = c.transaction(func(tx *ClusterTx) error {
			instances, err = tx.GetInstances(InstanceFilter{})
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get instances: %w", err)
		}

		for _, instance := range instances {
			if instance.ID != entityID {
				continue
			}

			uri = fmt.Sprintf(cluster.EntityURIs[entityType], instance.Name, instance.Project)
			break
		}

		if uri == "" {
			return "", ErrUnknownEntityID
		}
	case cluster.TypeInstanceBackup:
		instanceBackup, err := c.GetInstanceBackupWithID(entityID)
		if err != nil {
			return "", fmt.Errorf("Failed to get instance backup: %w", err)
		}

		var instances []Instance

		err = c.transaction(func(tx *ClusterTx) error {
			instances, err = tx.GetInstances(InstanceFilter{})
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get instances: %w", err)
		}

		for _, instance := range instances {
			if instance.ID != instanceBackup.InstanceID {
				continue
			}

			uri = fmt.Sprintf(cluster.EntityURIs[entityType], instance.Name, instanceBackup.Name, instance.Project)
			break
		}

		if uri == "" {
			return "", ErrUnknownEntityID
		}
	case cluster.TypeInstanceSnapshot:
		var snapshots []InstanceSnapshot

		err = c.transaction(func(tx *ClusterTx) error {
			snapshots, err = tx.GetInstanceSnapshots(InstanceSnapshotFilter{})
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get instance snapshots: %w", err)
		}

		for _, snapshot := range snapshots {
			if snapshot.ID != entityID {
				continue
			}

			uri = fmt.Sprintf(cluster.EntityURIs[entityType], snapshot.Name, snapshot.Project)
			break
		}

		if uri == "" {
			return "", ErrUnknownEntityID
		}
	case cluster.TypeNetwork:
		networkName, projectName, err := c.GetNetworkNameAndProjectWithID(entityID)
		if err != nil {
			return "", fmt.Errorf("Failed to get network name and project name: %w", err)
		}

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], networkName, projectName)
	case cluster.TypeNetworkACL:
		networkACLName, projectName, err := c.GetNetworkACLNameAndProjectWithID(entityID)
		if err != nil {
			return "", fmt.Errorf("Failed to get network ACL name and project name: %w", err)
		}

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], networkACLName, projectName)
	case cluster.TypeNode:
		var nodeInfo NodeInfo

		err := c.transaction(func(tx *ClusterTx) error {
			nodeInfo, err = tx.GetNodeWithID(entityID)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get node information: %w", err)
		}

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], nodeInfo.Name)
	case cluster.TypeOperation:
		var op Operation

		err = c.transaction(func(tx *ClusterTx) error {
			id := int64(entityID)
			filter := OperationFilter{ID: &id}
			ops, err := tx.GetOperations(filter)
			if err != nil {
				return err
			}

			if len(ops) > 1 {
				return fmt.Errorf("More than one operation matches")
			}

			op = ops[0]
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("Failed to get operation: %w", err)
		}

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], op.UUID)
	case cluster.TypeStoragePool:
		_, pool, _, err := c.GetStoragePoolWithID(entityID)
		if err != nil {
			return "", fmt.Errorf("Failed to get storage pool: %w", err)
		}

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], pool.Name)
	case cluster.TypeStorageVolume:
		args, err := c.GetStoragePoolVolumeWithID(entityID)
		if err != nil {
			return "", fmt.Errorf("Failed to get storage volume: %w", err)
		}

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], args.PoolName, args.TypeName, args.Name, args.ProjectName)
	case cluster.TypeStorageVolumeBackup:
		backup, err := c.GetStoragePoolVolumeBackupWithID(entityID)
		if err != nil {
			return "", fmt.Errorf("Failed to get volume backup: %w", err)
		}

		instance, err := c.GetStoragePoolVolumeWithID(int(backup.ID))
		if err != nil {
			return "", fmt.Errorf("Failed to get storage volume: %w", err)
		}

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], instance.PoolName, instance.Name, backup.Name, instance.ProjectName)
	case cluster.TypeStorageVolumeSnapshot:
		snapshot, err := c.GetStorageVolumeSnapshotWithID(entityID)
		if err != nil {
			return "", fmt.Errorf("Failed to get volume snapshot: %w", err)
		}

		fields := strings.Split(snapshot.Name, "/")

		uri = fmt.Sprintf(cluster.EntityURIs[entityType], snapshot.PoolName, snapshot, snapshot.TypeName, fields[0], fields[1], snapshot.ProjectName)
	}

	return uri, nil
}