---
layout: "pnap"
page_title: "phoenixNAP: pnap_rancher_cluster"
sidebar_current: "docs-pnap-datasource-rancher-cluster"
description: |-
  Provides a phoenixNAP Rancher Cluster datasource. This can be used to read Rancher Server deployment details.
---

# pnap_server Resource

Provides a phoenixNAP Rancher Cluster datasource. This can be used to read Rancher Server deployment details.



## Example Usage

Fetch a Rancher Cluster by ID or name and show it's details in alphabetical order. 

```hcl
# Fetch a Rancher Cluster
data "pnap_rancher_cluster" "test" {
  id = "123"
  name = "Rancher-Deployment-1"
}

# Show the Rancher Cluster details
output "rancher_cluster" {
  value = data.pnap_rancher_cluster.test
}
```

## Argument Reference

The following arguments are supported:

* `id` - The Cluster identifier.
* `name` - Cluster name.


## Attributes Reference

The following attributes are exported:

* `id` - The Cluster identifier.
* `name` - Cluster name.
* `description` - Cluster description.
* `location` - (Required) Deployment location. For a full list of available locations visit [API docs](https://developers.phoenixnap.com/docs/rancher/1)
* `initial_cluster_version` - The Rancher version that was installed on the cluster during the first creation process.
* `node_pools` - The node pools associated with the cluster (must contain exactly one item).
    * `node_pool` - Node Pool Configuration. A node pool contains the name and configuration for a cluster's node pool. Node pools are set of nodes with a common configuration and specification.
        * `name` - The name of the node pool.
        * `node_count` - Number of configured nodes. Currently only node counts of 1 and 3 are possible.
        * `server_type` - Node server type. Default value is "s0.d1.small". For a full list of allowed values visit [API docs](https://developers.phoenixnap.com/docs/rancher/1)
        * `ssh_config` - Configuration defining which public SSH keys are pre-installed as authorized on the server.
            * `install_default-keys` - Define whether public keys marked as default should be installed on this node. Default value is true.
            * `keys` - List of public SSH keys.
            * `key-ids` - List of public SSH key identifiers.
        * `nodes` - The nodes associated with this node pool.
            * `node` - Node details.
                * `server_id` - The server identifier.
* `metadata` - Connection parameters to use to connect to the Rancher Server Administrative GUI.
    * `url` - The Rancher Server URL.
    * `username` - The username to use to login to the Rancher Server. This field is returned only as a response to the create cluster request. Make sure to take note or you will not be able to access the server.
    * `password` - This is the password to be used to login to the Rancher Server. This field is returned only as a response to the create cluster request. Make sure to take note or you will not be able to access the server.
* `status_description` - The cluster status.
