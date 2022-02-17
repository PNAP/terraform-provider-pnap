package pnap

import (
	"fmt"
	"log"

	"github.com/PNAP/go-sdk-helper-bmc/command/ranchersolutionapi/cluster"
	"github.com/PNAP/go-sdk-helper-bmc/receiver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rancherapiclient "github.com/phoenixnap/go-sdk-bmc/ranchersolutionapi"
)

func resourceRancherCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceRancherClusterCreate,
		Read:   resourceRancherClusterRead,
		Update: resourceRancherClusterUpdate,
		Delete: resourceRancherClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(pnapRetryTimeout),
			Update: schema.DefaultTimeout(pnapRetryTimeout),
			Delete: schema.DefaultTimeout(pnapDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"initial_cluster_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_pools": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_pool": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"node_count": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"server_type": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"ssh_config": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"install_default_keys": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												"keys": {
													Type:     schema.TypeSet,
													Optional: true,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"key_ids": {
													Type:     schema.TypeSet,
													Optional: true,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"nodes": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"node": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"server_id": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"token": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"tls_san": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"etcd_snapshot_schedule_cron": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"etcd_snapshot_retention": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  5,
						},
						"node_taint": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cluster_domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"certificates": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ca_certificate": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"certificate": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"certificate_key": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"password": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
			},
			"status_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRancherClusterCreate(d *schema.ResourceData, m interface{}) error {

	client := m.(receiver.BMCSDK)

	request := &rancherapiclient.Cluster{}
	var name = d.Get("name").(string)
	if len(name) > 0 {
		request.Name = &name
	}
	var desc = d.Get("description").(string)
	if len(desc) > 0 {
		request.Description = &desc
	}
	request.Location = d.Get("location").(string)
	// node_pools block
	var nodePools = d.Get("node_pools").([]interface{})
	if len(nodePools) > 0 {

		pools := make([]rancherapiclient.NodePool, len(nodePools))

		for k, j := range nodePools {
			nodePoolObject := rancherapiclient.NodePool{}

			nodePoolsItem := j.(map[string]interface{})

			nodePool := nodePoolsItem["node_pool"].([]interface{})[0]
			nodePoolItem := nodePool.(map[string]interface{})

			name := nodePoolItem["name"].(string)
			nodeCount := int32(nodePoolItem["node_count"].(int))
			serverType := nodePoolItem["server_type"].(string)

			if len(name) > 0 {
				nodePoolObject.Name = &name
			}
			if nodeCount > 0 {
				nodePoolObject.NodeCount = &nodeCount
			}
			if len(serverType) > 0 {
				nodePoolObject.ServerType = &serverType
			}

			if nodePoolItem["ssh_config"] != nil && len(nodePoolItem["ssh_config"].([]interface{})) > 0 {
				sshConfigObject := rancherapiclient.SshConfig{}
				nodePoolObject.SshConfig = &sshConfigObject

				sshConfig := nodePoolItem["ssh_config"].([]interface{})[0]
				sshConfigItem := sshConfig.(map[string]interface{})

				installDefaultKeys := sshConfigItem["install_default_keys"].(bool)
				sshConfigObject.InstallDefaultKeys = &installDefaultKeys

				tempKeys := sshConfigItem["keys"].(*schema.Set).List()
				keys := make([]string, len(tempKeys))
				for i, v := range tempKeys {
					keys[i] = fmt.Sprint(v)
				}
				if len(keys) > 0 {
					sshConfigObject.Keys = &keys
				}

				tempKeyIds := sshConfigItem["key_ids"].(*schema.Set).List()
				keyIds := make([]string, len(tempKeyIds))
				for i, v := range tempKeyIds {
					keyIds[i] = fmt.Sprint(v)
				}
				if len(keyIds) > 0 {
					sshConfigObject.KeyIds = &keyIds
				}

			}
			pools[k] = nodePoolObject
		}
		request.NodePools = &pools
	}
	// end of node_pools block
	if d.Get("configuration") != nil && len(d.Get("configuration").([]interface{})) > 0 {
		configuration := d.Get("configuration").([]interface{})[0]
		configurationItem := configuration.(map[string]interface{})

		configurationObject := rancherapiclient.RancherClusterConfig{}

		token := configurationItem["token"].(string)
		if len(token) > 0 {
			configurationObject.Token = &token
		}
		tlsSan := configurationItem["tls_san"].(string)
		if len(tlsSan) > 0 {
			configurationObject.TlsSan = &tlsSan
		}
		etcdCron := configurationItem["etcd_snapshot_schedule_cron"].(string)
		if len(etcdCron) > 0 {
			configurationObject.EtcdSnapshotScheduleCron = &etcdCron
		}
		etcdRet := int32(configurationItem["etcd_snapshot_retention"].(int))
		configurationObject.EtcdSnapshotRetention = &etcdRet
		nodeTaint := configurationItem["node_taint"].(string)
		if len(nodeTaint) > 0 {
			configurationObject.NodeTaint = &nodeTaint
		}
		clustDom := configurationItem["cluster_domain"].(string)
		if len(clustDom) > 0 {
			configurationObject.ClusterDomain = &clustDom
		}
		if d.Get("certificates") != nil && len(d.Get("certificates").([]interface{})) > 0 {
			certificates := d.Get("certificates").([]interface{})[0]
			certificatesItem := certificates.(map[string]interface{})

			caCert := certificatesItem["ca_certificate"].(string)
			cert := certificatesItem["certificate"].(string)
			certKey := certificatesItem["certificate_key"].(string)

			if len(caCert) > 0 || len(cert) > 0 || len(certKey) > 0 {
				certificatesObject := rancherapiclient.RancherClusterCertificates{}
				configurationObject.Certificates = &certificatesObject

				if len(caCert) > 0 {
					certificatesObject.CaCertificate = &caCert
				}
				if len(cert) > 0 {
					certificatesObject.Certificate = &cert
				}
				if len(certKey) > 0 {
					certificatesObject.CertificateKey = &certKey
				}
			}
		}
		request.Configuration = &configurationObject
	}
	var statDesc = d.Get("status_description").(string)
	if len(statDesc) > 0 {
		request.StatusDescription = &statDesc
	}

	requestCommand := cluster.NewCreateClusterCommand(client, *request)
	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	} else if resp.Id == nil {
		return fmt.Errorf("unknown cluster identifier")
	} else {
		d.SetId(*resp.Id)

		waitResultError := clusterWaitForCreate(*resp.Id, &client)
		if waitResultError != nil {
			return waitResultError
		}
	}

	return resourceRancherClusterRead(d, m)
}

func resourceRancherClusterRead(d *schema.ResourceData, m interface{}) error {
	client := m.(receiver.BMCSDK)
	clusterID := d.Id()

	requestCommand := cluster.NewGetClusterCommand(client, clusterID)
	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	}
	if resp.Id == nil {
		return fmt.Errorf("unknown cluster identifier")
	}
	d.SetId(*resp.Id)
	if resp.Name != nil {
		d.Set("name", *resp.Name)
	}
	if resp.Description != nil {
		d.Set("description", *resp.Description)
	}
	d.Set("location", resp.Location)
	if resp.InitialClusterVersion != nil {
		d.Set("initial_cluster_version", *resp.InitialClusterVersion)
	}
	if resp.NodePools != nil {
		nodePools := flattenNodePools(*resp.NodePools)
		if err := d.Set("node_pools", nodePools); err != nil {
			return err
		}
	}
	if resp.Metadata != nil {
		metaData := make(map[string]interface{})
		if resp.Metadata.Url != nil {
			metaData["url"] = *resp.Metadata.Url
		}
		if resp.Metadata.Username != nil {
			metaData["username"] = *resp.Metadata.Username
		}
		if resp.Metadata.Password != nil {
			metaData["password"] = *resp.Metadata.Password
		}
		d.Set("metadata", metaData)
	}
	if resp.StatusDescription != nil {
		d.Set("status_description", *resp.StatusDescription)
	}
	return nil
}

func resourceRancherClusterUpdate(d *schema.ResourceData, m interface{}) error {
	return fmt.Errorf("unsuported action")
}

func resourceRancherClusterDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(receiver.BMCSDK)
	clusterID := d.Id()

	requestCommand := cluster.NewDeleteClusterCommand(client, clusterID)
	_, err := requestCommand.Execute()
	if err != nil {
		return err
	}
	return nil
}

func flattenNodePools(nodePools []rancherapiclient.NodePool) []interface{} {
	if nodePools != nil {
		np := make([]interface{}, len(nodePools))
		for i, v := range nodePools {
			n := make(map[string]interface{})
			if v.Name != nil {
				n["name"] = *v.Name
			}
			if v.NodeCount != nil {
				n["node_count"] = int(*v.NodeCount)
			}
			if v.ServerType != nil {
				n["server_type"] = *v.ServerType
			}
			if v.SshConfig != nil {
				sc := make(map[string]interface{})
				if v.SshConfig.InstallDefaultKeys != nil {
					sc["install_default_keys"] = *v.SshConfig.InstallDefaultKeys
				}
				if v.SshConfig.Keys != nil {
					keys := make([]interface{}, len(*v.SshConfig.Keys))
					for j, k := range *v.SshConfig.Keys {
						keys[j] = k
					}
					sc["keys"] = keys
				}
				if v.SshConfig.KeyIds != nil {
					keyIds := make([]interface{}, len(*v.SshConfig.KeyIds))
					for j, k := range *v.SshConfig.KeyIds {
						keyIds[j] = k
					}
					sc["key_ids"] = keyIds
				}
				n["ssh_config"] = sc
			}
			if v.Nodes != nil {
				nodes := make([]interface{}, len(*v.Nodes))
				for j, k := range *v.Nodes {
					node := make(map[string]interface{})
					if k.ServerId != nil {
						node["server_id"] = *k.ServerId
					}
					nodes[j] = node
				}
				n["nodes"] = nodes
			}
			np[i] = n
		}
		return np
	}
	return make([]interface{}, 0)
}

func clusterWaitForCreate(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for cluster %s to be created...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Creating"},
		Target:     []string{"Ready", "Error"},
		Refresh:    clusterRefreshForCreate(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for cluster (%s) to switch to target state: %v", id, err)
	}

	return nil
}

func clusterRefreshForCreate(client *receiver.BMCSDK, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		requestCommand := cluster.NewGetClusterCommand(*client, id)

		resp, err := requestCommand.Execute()
		if err != nil {
			return 0, "", err
		} else if resp.StatusDescription != nil {
			return 0, *resp.StatusDescription, nil
		} else {
			return 0, "", nil
		}
	}
}
