# Kubernetes NFS Storage Setup for SHARK Team

This guide explains how to set up shared NFS (Network File System) storage in our OCI Kubernetes cluster. 
This storage solution allows all team members to access and share data across their Kubernetes jobs.

## Overview

The setup provides:
- Shared storage accessible by all team members
- Persistent data storage that survives pod restarts
- Ability to run multiple jobs accessing the same data simultaneously
- Consistent storage solution across all Kubernetes jobs

## Prerequisites

- Access to the OCI Kubernetes cluster
- NFS server details (will be provided by infrastructure team)
- `kubectl` configured to access the cluster

## Quick Start

1. Clone this repository and navigate to the storage configuration directory:
   ```bash
   git clone <repository-url>
   cd k8s-storage-config
   ```

2. Update the configuration in `nfs-config.yaml`:
   - Replace `your-nfs-server.internal` with the actual NFS server address
   - Adjust storage size if needed (default 100Gi)
   - Modify the NFS path if required (default "/shared")

3. Apply the configuration:
   ```bash
   kubectl apply -f nfs-config.yaml
   ```

## Usage Example

To use the NFS storage in your jobs:

1. Reference the PVC in your job YAML:
   ```yaml
   volumes:
   - name: nfs-storage
     persistentVolumeClaim:
       claimName: nfs-pvc
   ```

2. Mount the volume in your container:
   ```yaml
   volumeMounts:
   - name: nfs-storage
     mountPath: /mnt/nfs
   ```

## Common Operations

### Checking Storage Status
```bash
# Check PV status
kubectl get pv

# Check PVC status
kubectl get pvc

# Check mounted volumes in a pod
kubectl describe pod <pod-name>
```

### Data Persistence
- Data written to `/mnt/nfs` (or your chosen mount path) persists across pod restarts
- Multiple pods can read/write simultaneously
- Data remains even after pods or jobs complete

## Best Practices

1. **File Organization**
   - Create team-specific subdirectories for better organization
   - Use clear naming conventions for shared files
   - Clean up unnecessary data regularly

2. **Resource Management**
   - Monitor storage usage to avoid running out of space
   - Use appropriate file permissions

3. **Performance**
   - Avoid storing too many small files
   - Consider data locality when processing large datasets

## Troubleshooting

Common issues and solutions:

1. **Unable to mount volume**
   - Check NFS server accessibility
   - Verify network policies allow NFS traffic
   - Ensure PV and PVC are properly bound

2. **Permission issues**
   - Check file permissions on NFS server
   - Verify pod security contexts

## Support

For issues or questions:
- Create a ticket in [your-ticket-system]
- Contact the infrastructure team
- Check Kubernetes logs: `kubectl logs <pod-name>`

## Future Improvements

Planned enhancements:
- Storage quota implementation per team/project
- Automated backup solution
- Storage monitoring and alerts
- Data cleanup policies

---
Note: This is a living document. Please contribute improvements or report issues via Teams/Slack.
