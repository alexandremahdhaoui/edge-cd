# edgectl - E2E tests

### Prerequisites

##### qemu-kvm & libvirt

```bash
sudo apt update
sudo apt install qemu-kvm libvirt-daemon-system # libvirt-dev pkg-config
sudo systemctl enable --now libvirtd
sudo usermod -a -G libvirt "${USER}"
# -- verify group by running:
# getent group | grep libvirt
# -- test by running
virsh --connect qemu:///system list --all
```

##### qemu-img

```bash
sudo apt-get update 
sudo apt-get install xorriso # genisoimage
```

### Virsh debugging

```bash
virsh net-dumpxml default | yq -Pp xml
virsh net-dhcp-release default
```

### Clean up

```bash
{
    read -rp "Enter VM name: " VM_NAME
    virsh destroy "${VM_NAME}"
    virsh undefine --remove-all-storage "${VM_NAME}" 
    sudo rm -f "/tmp/${VM_NAME}*"
}
```
