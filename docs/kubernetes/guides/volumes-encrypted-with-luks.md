<!--
---
date: "2026-09-07"
date_changed: "2026-10-21"
title: "Encrypting Volumes with LUKS"
tags: []
language: "en"
description: ""
docs_type: ["how_to"]
product_category: ["Integrations"]
translation: ["Integrations", "CSI driver", "Guides", "Encrypting Volumes with LUKS"]
scrape_type: "whole"
priority: 80
---
-->

# Volumes Encrypted with LUKS

To add encryption with LUKS you have to create a dedicate secret containing an encryption passphrase and duplicate the default `hcloud-volumes` storage class with added parameters referencing this secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: encryption-secret
  namespace: kube-system
stringData:
  encryption-passphrase: foobar

---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: hcloud-volumes-encrypted
provisioner: csi.hetzner.cloud
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
parameters:
  csi.storage.k8s.io/node-publish-secret-name: encryption-secret
  csi.storage.k8s.io/node-publish-secret-namespace: kube-system
```

Your nodes might need to have `cryptsetup` installed to mount the Volumes with LUKS.

## Disabling the dm-crypt workqueues

By default, dm-crypt routes every read and write of an encrypted Volume through internal kernel workqueues. On Hetzner Cloud Volumes this adds queueing latency, most visible in the tail (p99) of write and `fsync` latency.

You can open encrypted Volumes without these workqueues by adding `encryption-disable-workqueues: "true"` to the encryption secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: encryption-secret
  namespace: kube-system
stringData:
  encryption-passphrase: foobar
  encryption-disable-workqueues: "true"
```

The driver then passes `--perf-no_read_workqueue --perf-no_write_workqueue` to `cryptsetup luksOpen`. This requires cryptsetup 2.3.4 or newer and Linux 5.9 or newer on the node.

The option only takes effect when the driver opens the LUKS device, so Volumes that are already attached keep their current settings until they are detached and attached again, for example when the Pod moves to another node.

