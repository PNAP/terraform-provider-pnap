package pnap

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/PNAP/go-sdk-helper-bmc/command/networkapi/publicnetwork"
	"github.com/PNAP/go-sdk-helper-bmc/receiver"

	networkapiclient "github.com/phoenixnap/go-sdk-bmc/networkapi"
)

func resourcePublicNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourcePublicNetworkCreate,
		Read:   resourcePublicNetworkRead,
		Update: resourcePublicNetworkUpdate,
		Delete: resourcePublicNetworkDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(pnapRetryTimeout),
			Update: schema.DefaultTimeout(pnapRetryTimeout),
			Delete: schema.DefaultTimeout(pnapDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ip_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"public_network_ip_block": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"created_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vlan_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memberships": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourcePublicNetworkCreate(d *schema.ResourceData, m interface{}) error {

	client := m.(receiver.BMCSDK)

	request := &networkapiclient.PublicNetworkCreate{}
	request.Name = d.Get("name").(string)
	request.Location = d.Get("location").(string)
	var desc = d.Get("description").(string)
	if len(desc) > 0 {
		request.Description = &desc
	}
	ipBlocks := d.Get("ip_blocks").([]interface{})
	if len(ipBlocks) > 0 {
		ipBlocksObject := make([]networkapiclient.PublicNetworkIpBlock, len(ipBlocks))
		for i, j := range ipBlocks {
			ibItem := j.(map[string]interface{})
			pnib := ibItem["public_network_ip_block"].([]interface{})[0]
			pnibItem := pnib.(map[string]interface{})

			pnibObject := networkapiclient.PublicNetworkIpBlock{}
			pnibObject.Id = pnibItem["id"].(string)
			ipBlocksObject[i] = pnibObject
		}
		request.IpBlocks = &ipBlocksObject
	}
	requestCommand := publicnetwork.NewCreatePublicNetworkCommand(client, *request)

	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	}

	d.SetId(resp.Id)

	return resourcePublicNetworkRead(d, m)
}

func resourcePublicNetworkRead(d *schema.ResourceData, m interface{}) error {
	client := m.(receiver.BMCSDK)
	networkID := d.Id()
	requestCommand := publicnetwork.NewGetPublicNetworkCommand(client, networkID)
	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	}
	d.SetId(resp.Id)
	d.Set("name", resp.Name)
	d.Set("location", resp.Location)
	desc := resp.Description
	if desc != nil {
		d.Set("description", *resp.Description)
	} else {
		d.Set("description", "")
	}

	ibInput := d.Get("ip_blocks").([]interface{})
	ipBlocks := flattenIpBlocks(resp.IpBlocks, ibInput)

	if err := d.Set("ip_blocks", ipBlocks); err != nil {
		return err
	}

	if len(resp.CreatedOn.String()) > 0 {
		d.Set("created_on", resp.CreatedOn.String())
	}
	d.Set("vlan_id", resp.VlanId)

	memberships := flattenMemberships(resp.Memberships)

	if err := d.Set("memberships", memberships); err != nil {
		return err
	}
	return nil
}

func resourcePublicNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	if d.HasChange("name") || d.HasChange("description") {
		client := m.(receiver.BMCSDK)
		networkID := d.Id()
		request := &networkapiclient.PublicNetworkModify{}
		var name = d.Get("name").(string)
		request.Name = &name
		var desc = d.Get("description").(string)
		request.Description = &desc
		requestCommand := publicnetwork.NewUpdatePublicNetworkCommand(client, networkID, *request)
		_, err := requestCommand.Execute()
		if err != nil {
			return err
		}
	} else if d.HasChange("ip_blocks") {
		client := m.(receiver.BMCSDK)
		networkID := d.Id()
		requestCommand := publicnetwork.NewGetPublicNetworkCommand(client, networkID)
		resp, err := requestCommand.Execute()
		if err != nil {
			return err
		}
		ipBlocks := resp.IpBlocks
		ipBlocksInput := d.Get("ip_blocks").([]interface{})

		var sameIpBlocks []interface{}

		if len(ipBlocks) > 0 && ipBlocksInput != nil && len(ipBlocksInput) > 0 {
			for _, j := range ipBlocks {
				id := j.Id
				for _, l := range ipBlocksInput {
					ipbsInputItem := l.(map[string]interface{})
					if ipbsInputItem["public_network_ip_block"] != nil && len(ipbsInputItem["public_network_ip_block"].([]interface{})) > 0 {
						pnibInput := ipbsInputItem["public_network_ip_block"].([]interface{})[0]
						pnibInputItem := pnibInput.(map[string]interface{})
						idInput := pnibInputItem["id"].(string)
						if id == idInput {
							sameIpBlocks = append(sameIpBlocks, id)
						}
					}
				}
			}
		}
		if len(ipBlocks) > len(sameIpBlocks) {
			for _, j := range ipBlocks {
				id := j.Id
				var same = false
				for _, l := range sameIpBlocks {
					if id == l {
						same = true
					}
				}
				if !same {
					requestCommand := publicnetwork.NewRemoveIpBlockFromPublicNetworkCommand(client, networkID, id)
					_, err := requestCommand.Execute()
					if err != nil {
						return err
					}
				}
			}
		} else if len(ipBlocksInput) > len(sameIpBlocks) {
			for _, l := range ipBlocksInput {
				ipbsInputItem := l.(map[string]interface{})
				if ipbsInputItem["public_network_ip_block"] != nil && len(ipbsInputItem["public_network_ip_block"].([]interface{})) > 0 {
					pnibInput := ipbsInputItem["public_network_ip_block"].([]interface{})[0]
					pnibInputItem := pnibInput.(map[string]interface{})
					idInput := pnibInputItem["id"].(string)
					var same = false
					for _, l := range sameIpBlocks {
						if idInput == l {
							same = true
						}
					}
					if !same {
						pnibObject := networkapiclient.PublicNetworkIpBlock{}
						pnibObject.Id = idInput
						request := &pnibObject
						requestCommand := publicnetwork.NewAddIpBlock2PublicNetworkCommand(client, networkID, *request)
						_, err := requestCommand.Execute()
						if err != nil {
							return err
						}
					}
				}
			}
		}
	} else {
		return fmt.Errorf("unsupported action")
	}
	return resourcePublicNetworkRead(d, m)
}

func resourcePublicNetworkDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(receiver.BMCSDK)

	networkID := d.Id()

	requestCommand := publicnetwork.NewDeletePublicNetworkCommand(client, networkID)
	err := requestCommand.Execute()
	if err != nil {
		return err
	}

	return nil
}

func flattenMemberships(memberships []networkapiclient.PublicNetworkMembership) []interface{} {
	if memberships != nil {
		mems := make([]interface{}, len(memberships))
		for i, v := range memberships {
			mem := make(map[string]interface{})
			mem["resource_id"] = v.ResourceId
			mem["resource_type"] = v.ResourceType
			ips := make([]interface{}, len(v.Ips))
			for j, k := range v.Ips {
				ips[j] = k
			}
			mem["ips"] = ips
			mems[i] = mem
		}
		return mems
	}
	return make([]interface{}, 0)
}

func flattenIpBlocks(ipBlocks []networkapiclient.PublicNetworkIpBlock, ibInput []interface{}) []interface{} {
	if len(ipBlocks) > 0 {
		var ipb []interface{}
		if len(ibInput) == 0 || ibInput[0] == nil {
			for _, j := range ipBlocks {
				ipbItem := make(map[string]interface{})
				pnipb := make([]interface{}, 1)
				pnipbItem := make(map[string]interface{})

				pnipbItem["id"] = j.Id

				pnipb[0] = pnipbItem
				ipbItem["public_network_ip_block"] = pnipb
				ipb = append(ipb, ipbItem)
			}
		} else if len(ibInput) > 0 {
			for i := range ibInput {
				for _, l := range ipBlocks {
					if ibInput[i].(map[string]interface{})["public_network_ip_block"].([]interface{})[0].(map[string]interface{})["id"] == l.Id {
						ipbItem := make(map[string]interface{})
						pnipb := make([]interface{}, 1)
						pnipbItem := make(map[string]interface{})

						pnipbItem["id"] = l.Id

						pnipb[0] = pnipbItem
						ipbItem["public_network_ip_block"] = pnipb
						ipb = append(ipb, ipbItem)
					}
				}
			}
			for _, p := range ipBlocks {
				var newIpBlock = true
				for r := range ipb {
					if p.Id == ipb[r].(map[string]interface{})["public_network_ip_block"].([]interface{})[0].(map[string]interface{})["id"] {
						newIpBlock = false
					}
				}
				if newIpBlock {
					ipbItem := make(map[string]interface{})
					pnipb := make([]interface{}, 1)
					pnipbItem := make(map[string]interface{})

					pnipbItem["id"] = p.Id

					pnipb[0] = pnipbItem
					ipbItem["public_network_ip_block"] = pnipb
					ipb = append(ipb, ipbItem)
				}
			}
		}
		return ipb
	}
	return make([]interface{}, 0)
}
