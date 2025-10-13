# AgentaFlow AWS Demo Troubleshooting Guide

**Comprehensive troubleshooting guide for AgentaFlow SRO Community Edition AWS demo deployment**

This guide covers common issues encountered during AWS demo setup and provides step-by-step solutions to get your demo running smoothly.

## 🚨 Quick Diagnosis

### **Before You Start**
Run this quick diagnostic to identify potential issues:

```bash
#!/bin/bash
# quick-diagnosis.sh

echo "=== AgentaFlow Demo Environment Check ==="

# Check AWS CLI configuration
echo "Checking AWS CLI..."
aws sts get-caller-identity || echo "❌ AWS CLI not configured properly"

# Check required tools
echo "Checking required tools..."
kubectl version --client || echo "❌ kubectl not installed"
eksctl version || echo "❌ eksctl not installed" 
helm version || echo "❌ helm not installed"

# Check AWS permissions
echo "Checking AWS permissions..."
aws iam list-attached-role-policies --role-name eksctl-agentaflow-demo-cluster-ServiceRole 2>/dev/null || echo "⚠️  EKS permissions may be insufficient"

# Check region and availability zones
echo "Checking GPU instance availability..."
aws ec2 describe-instance-type-offerings --location-type availability-zone \
  --filters Name=instance-type,Values=g4dn.xlarge --region us-west-2 \
  --query 'InstanceTypeOfferings[*].Location' --output text || echo "⚠️  GPU instances may not be available in region"

echo "=== Diagnosis Complete ==="
```

---

## 🛠️ Prerequisites Issues

### **Issue: AWS CLI Not Configured**
```
Error: Unable to locate credentials
```

#### **Solution:**
```bash
# Configure AWS CLI with your credentials
aws configure

# Verify configuration
aws sts get-caller-identity

# For demo, ensure you have these minimum permissions:
# - EC2: Full access for instance management
# - EKS: Full access for cluster creation
# - IAM: Permissions to create service roles
# - VPC: Permissions to create networking resources
```

#### **Alternative: Using AWS Profile**
```bash
# Create a demo-specific profile
aws configure --profile agentaflow-demo

# Use profile for all demo commands
export AWS_PROFILE=agentaflow-demo

# Verify profile works
aws sts get-caller-identity --profile agentaflow-demo
```

### **Issue: Insufficient AWS Permissions**
```
Error: AccessDenied when calling CreateCluster
```

#### **Required IAM Policies:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "eks:*",
        "ec2:*",
        "iam:CreateRole",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy",
        "iam:DeleteRole",
        "iam:ListAttachedRolePolicies",
        "iam:CreateInstanceProfile",
        "iam:DeleteInstanceProfile",
        "iam:AddRoleToInstanceProfile",
        "iam:RemoveRoleFromInstanceProfile",
        "iam:PassRole",
        "cloudformation:*"
      ],
      "Resource": "*"
    }
  ]
}
```

#### **Quick Permission Check:**
```bash
# Test key permissions
aws eks list-clusters --region us-west-2
aws ec2 describe-regions
aws iam list-roles --max-items 1

# If any fail, contact your AWS administrator
```

### **Issue: Missing Tools**
```
Command 'eksctl' not found
Command 'kubectl' not found  
Command 'helm' not found
```

#### **Installation Solutions:**

**Windows (PowerShell):**
```powershell
# Install Chocolatey if not present
Set-ExecutionPolicy Bypass -Scope Process -Force
iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))

# Install required tools
choco install kubernetes-cli
choco install eksctl
choco install kubernetes-helm

# Verify installations
kubectl version --client
eksctl version
helm version
```

**macOS:**
```bash
# Install Homebrew if not present
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install required tools
brew install kubectl eksctl helm

# Verify installations
kubectl version --client
eksctl version  
helm version
```

**Linux:**
```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install eksctl
curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin

# Install helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Verify installations
kubectl version --client
eksctl version
helm version
```

---

## 🏗️ EKS Cluster Creation Issues

### **Issue: Spot Instance Unavailable**
```
Error: no spot instance pools available for the requirements
```

#### **Solutions:**

**Option 1: Use Multiple Instance Types**
```bash
# Modify eksctl command to include multiple instance types
eksctl create nodegroup \
  --cluster agentaflow-demo \
  --region us-west-2 \
  --name gpu-workers \
  --node-type g4dn.xlarge \
  --nodes 2 \
  --nodes-min 1 \
  --nodes-max 5 \
  --spot \
  --instance-types g4dn.xlarge,g4dn.2xlarge,g5.xlarge \
  --ssh-access \
  --ssh-public-key your-key-name
```

**Option 2: Use On-Demand Instances**
```bash
# Remove --spot flag for guaranteed availability
eksctl create nodegroup \
  --cluster agentaflow-demo \
  --region us-west-2 \
  --name gpu-workers \
  --node-type g4dn.xlarge \
  --nodes 2 \
  --nodes-min 1 \
  --nodes-max 5 \
  --ssh-access \
  --ssh-public-key your-key-name
```

**Option 3: Try Different Regions**
```bash
# Check GPU availability in different regions
for region in us-west-2 us-east-1 eu-west-1; do
  echo "Checking $region..."
  aws ec2 describe-instance-type-offerings \
    --location-type availability-zone \
    --filters Name=instance-type,Values=g4dn.xlarge \
    --region $region \
    --query 'InstanceTypeOfferings[*].Location' \
    --output text
done
```

### **Issue: VPC/Subnet Limits Exceeded**
```
Error: The maximum number of VPCs has been reached
```

#### **Solution: Use Existing VPC**
```bash
# List existing VPCs
aws ec2 describe-vpcs --region us-west-2

# Create cluster in existing VPC
eksctl create cluster \
  --name agentaflow-demo \
  --region us-west-2 \
  --vpc-private-subnets subnet-12345,subnet-67890 \
  --vpc-public-subnets subnet-abcde,subnet-fghij \
  --version 1.28
```

#### **Alternative: Clean Up Unused VPCs**
```bash
# List VPCs and their usage
aws ec2 describe-vpcs --region us-west-2 \
  --query 'Vpcs[*].{VpcId:VpcId,IsDefault:IsDefault,CidrBlock:CidrBlock}'

# Delete unused VPCs (be careful!)
# aws ec2 delete-vpc --vpc-id vpc-unused-id
```

### **Issue: Service Limits Exceeded**
```
Error: You have requested more instances than your current instance limit
```

#### **Check Current Limits:**
```bash
# Check EC2 service limits
aws service-quotas get-service-quota \
  --service-code ec2 \
  --quota-code L-DB2E81BA \
  --region us-west-2

# Check specific instance limits  
aws ec2 describe-account-attributes \
  --attribute-names max-instances \
  --region us-west-2
```

#### **Request Limit Increase:**
```bash
# Request GPU instance limit increase
aws service-quotas request-service-quota-increase \
  --service-code ec2 \
  --quota-code L-DB2E81BA \
  --desired-value 10 \
  --region us-west-2
```

### **Issue: SSH Key Not Found**
```
Error: InvalidKeyPair.NotFound
```

#### **Create SSH Key:**
```bash
# Create new key pair
aws ec2 create-key-pair \
  --key-name agentaflow-demo-key \
  --region us-west-2 \
  --query 'KeyMaterial' \
  --output text > ~/.ssh/agentaflow-demo-key.pem

# Set proper permissions
chmod 400 ~/.ssh/agentaflow-demo-key.pem

# Use in eksctl command
--ssh-public-key agentaflow-demo-key
```

---

## 🎮 NVIDIA GPU Operator Issues

### **Issue: GPU Operator Pods Stuck in Pending**
```
gpu-operator-node-feature-discovery pods stuck in Pending state
```

#### **Diagnosis:**
```bash
# Check node labels and taints
kubectl get nodes --show-labels
kubectl describe nodes -l node.kubernetes.io/instance-type=g4dn.xlarge

# Check GPU operator pods
kubectl get pods -n gpu-operator
kubectl describe pods -n gpu-operator
```

#### **Common Solutions:**

**Taint Issues:**
```bash
# Remove problematic taints
kubectl taint nodes --all node.kubernetes.io/not-ready-
kubectl taint nodes --all node.kubernetes.io/unreachable-

# Check if GPU nodes have special taints
kubectl describe node -l nvidia.com/gpu.present=true
```

**Resource Constraints:**
```bash
# Check node resources
kubectl top nodes
kubectl describe node -l node.kubernetes.io/instance-type=g4dn.xlarge

# Scale up if needed
eksctl scale nodegroup --cluster agentaflow-demo --name gpu-workers --nodes 3
```

### **Issue: GPU Operator Installation Fails**
```
Error: INSTALLATION FAILED: failed to download "nvidia/gpu-operator"
```

#### **Solution:**
```bash
# Update Helm repositories
helm repo update

# Verify NVIDIA repo
helm repo list | grep nvidia || helm repo add nvidia https://helm.ngc.nvidia.com/nvidia

# Try alternative installation method
curl -fsSL -o gpu-operator-v23.9.1.tgz \
  https://helm.ngc.nvidia.com/nvidia/gpu-operator/charts/gpu-operator-v23.9.1.tgz
helm install gpu-operator gpu-operator-v23.9.1.tgz \
  --namespace gpu-operator \
  --create-namespace \
  --set driver.enabled=false \
  --set toolkit.enabled=false
```

### **Issue: GPUs Not Detected**
```
No GPUs found on nodes
```

#### **Verification Steps:**
```bash
# Check if nodes have GPU hardware
kubectl get nodes -l node.kubernetes.io/instance-type=g4dn.xlarge -o wide

# Check node GPU labels
kubectl get nodes -l nvidia.com/gpu.present=true

# Verify GPU device plugin
kubectl get daemonset -n gpu-operator
kubectl logs -n gpu-operator -l app=nvidia-device-plugin-daemonset
```

#### **Force GPU Detection:**
```bash
# Restart GPU operator components
kubectl rollout restart daemonset/nvidia-device-plugin-daemonset -n gpu-operator
kubectl rollout restart daemonset/gpu-feature-discovery -n gpu-operator

# Check GPU resources
kubectl describe node -l nvidia.com/gpu.present=true | grep nvidia.com/gpu
```

---

## 🚀 AgentaFlow Deployment Issues

### **Issue: AgentaFlow Scheduler Not Starting**
```
scheduler-deployment pods CrashLoopBackOff
```

#### **Diagnosis:**
```bash
# Check scheduler logs
kubectl logs -n agentaflow-system deployment/agentaflow-scheduler

# Check configuration
kubectl describe configmap -n agentaflow-system
kubectl describe secret -n agentaflow-system

# Check RBAC permissions
kubectl auth can-i create pods --as=system:serviceaccount:agentaflow-system:agentaflow-scheduler
```

#### **Common Fixes:**

**Missing RBAC:**
```yaml
# Apply comprehensive RBAC
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: agentaflow-scheduler
rules:
- apiGroups: [""]
  resources: ["pods", "nodes"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: agentaflow-scheduler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: agentaflow-scheduler
subjects:
- kind: ServiceAccount
  name: agentaflow-scheduler
  namespace: agentaflow-system
```

**Configuration Issues:**
```bash
# Check if namespace exists
kubectl create namespace agentaflow-system --dry-run=client -o yaml | kubectl apply -f -

# Verify service account
kubectl get serviceaccount -n agentaflow-system agentaflow-scheduler || \
kubectl create serviceaccount agentaflow-scheduler -n agentaflow-system
```

### **Issue: Workloads Not Being Scheduled**
```
Pods remain in Pending state with AgentaFlow labels
```

#### **Debugging Steps:**
```bash
# Check scheduler status
kubectl get pods -n agentaflow-system -l app=agentaflow-scheduler

# Check events
kubectl get events --field-selector reason=FailedScheduling

# Verify GPU resources
kubectl describe nodes -l nvidia.com/gpu.present=true | grep -A 5 -B 5 nvidia.com/gpu

# Check AgentaFlow logs for scheduling decisions
kubectl logs -n agentaflow-system -l app=agentaflow-scheduler --tail=100
```

#### **Force Reschedule:**
```bash
# Delete stuck pods to trigger rescheduling
kubectl delete pods -l agentaflow.gpu/managed=true --field-selector=status.phase=Pending

# Restart scheduler
kubectl rollout restart deployment/agentaflow-scheduler -n agentaflow-system
```

---

## 📊 Monitoring Stack Issues

### **Issue: Prometheus Not Collecting Metrics**
```
No data available in Grafana dashboards
```

#### **Check Prometheus Status:**
```bash
# Verify Prometheus pods
kubectl get pods -n monitoring -l app.kubernetes.io/name=prometheus

# Check Prometheus configuration
kubectl logs -n monitoring -l app.kubernetes.io/name=prometheus --tail=50

# Verify service discovery
kubectl get servicemonitor -n monitoring
kubectl get endpoints -n monitoring
```

#### **Fix Service Discovery:**
```bash
# Ensure AgentaFlow metrics service exists
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: agentaflow-metrics
  namespace: agentaflow-system
  labels:
    app: agentaflow-scheduler
spec:
  ports:
  - port: 8080
    targetPort: 8080
    name: metrics
  selector:
    app: agentaflow-scheduler
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: agentaflow-scheduler
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: agentaflow-scheduler
  endpoints:
  - port: metrics
    path: /metrics
  namespaceSelector:
    matchNames:
    - agentaflow-system
EOF
```

### **Issue: Grafana Dashboard Not Loading**
```
Dashboard shows "No data points" or fails to load
```

#### **Debug Dashboard Issues:**
```bash
# Access Grafana directly
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

# Check data sources
# Navigate to Configuration > Data Sources in Grafana UI
# Test Prometheus connection

# Import AgentaFlow dashboard manually
# Copy content from examples/monitoring/grafana-dashboards/agentaflow.json
```

#### **Manual Dashboard Creation:**
```json
{
  "dashboard": {
    "title": "AgentaFlow GPU Utilization",
    "panels": [
      {
        "title": "GPU Utilization",
        "type": "graph",
        "targets": [
          {
            "expr": "nvidia_gpu_utilization_percent",
            "legendFormat": "GPU {{gpu}} on {{instance}}"
          }
        ]
      }
    ]
  }
}
```

---

## 💰 Cost Tracking Issues

### **Issue: Cost Data Not Appearing**
```
AgentaFlow dashboard shows $0.00 costs
```

#### **Verify AWS Cost Configuration:**
```bash
# Check AWS credentials for cost access
aws ce get-cost-and-usage \
  --time-period Start=2025-01-01,End=2025-01-02 \
  --granularity DAILY \
  --metrics UnblendedCost

# Verify AgentaFlow cost configuration
kubectl get configmap agentaflow-config -n agentaflow-system -o yaml
```

#### **Fix Cost Tracking Configuration:**
```bash
# Update cost configuration
kubectl create configmap agentaflow-config \
  --from-literal=enable-cost-tracking=true \
  --from-literal=aws-region=us-west-2 \
  --from-literal=cost-per-gpu-hour-g4dn-xlarge=0.526 \
  --from-literal=cost-per-gpu-hour-g4dn-2xlarge=1.052 \
  --from-literal=cost-per-gpu-hour-g5-xlarge=1.006 \
  -n agentaflow-system \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart scheduler to pick up new config
kubectl rollout restart deployment/agentaflow-scheduler -n agentaflow-system
```

### **Issue: AWS Cost Explorer Access Denied**
```
Error: AccessDenied when calling GetCostAndUsage
```

#### **Required IAM Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ce:GetCostAndUsage",
        "ce:GetDimensionValues",
        "ce:GetReservationCoverage",
        "ce:GetReservationPurchaseRecommendation",
        "ce:GetReservationUtilization",
        "ce:ListCostCategoryDefinitions"
      ],
      "Resource": "*"
    }
  ]
}
```

---

## 🔄 Workload Issues

### **Issue: Demo Workloads Failing to Start**
```
PyTorch training jobs failing with ImagePullBackOff
```

#### **Check Image Availability:**
```bash
# Verify image exists and is accessible
docker pull pytorch/pytorch:latest
docker pull nvidia/cuda:11.8-runtime-ubuntu20.04

# Check node image pull status
kubectl describe pod -l app=pytorch-training
```

#### **Alternative Images:**
```yaml
# Use alternative PyTorch image
spec:
  containers:
  - name: trainer
    image: pytorch/pytorch:1.12.1-cuda11.3-cudnn8-runtime
    # or
    image: nvcr.io/nvidia/pytorch:22.08-py3
```

### **Issue: GPU Memory Allocation Errors**
```
CUDA out of memory errors in workload logs
```

#### **Diagnosis:**
```bash
# Check GPU memory usage
kubectl exec -it gpu-node -- nvidia-smi

# Check workload resource requests
kubectl describe pod -l agentaflow.gpu/managed=true | grep -A 10 -B 10 "nvidia.com/gpu"
```

#### **Adjust Resource Requests:**
```yaml
# Reduce GPU memory requirements
spec:
  containers:
  - name: trainer
    resources:
      limits:
        nvidia.com/gpu: 1
        memory: 4Gi  # Reduce from 8Gi
      requests:
        memory: 2Gi  # Reduce from 4Gi
    env:
    - name: AGENTAFLOW_MEMORY_REQUIRED
      value: "2048"  # Reduce from 6144
```

---

## 🌐 Networking Issues

### **Issue: Cannot Access Grafana Dashboard**
```
Connection refused when accessing http://localhost:3000
```

#### **Troubleshoot Port Forwarding:**
```bash
# Check if pods are running
kubectl get pods -n monitoring -l app.kubernetes.io/name=grafana

# Try different port forwarding approach
kubectl port-forward -n monitoring deployment/prometheus-grafana 3000:3000

# Check service configuration
kubectl get svc -n monitoring -l app.kubernetes.io/name=grafana
kubectl describe svc prometheus-grafana -n monitoring
```

#### **Alternative Access Methods:**
```bash
# Use LoadBalancer service (costs extra)
kubectl patch svc prometheus-grafana -n monitoring -p '{"spec":{"type":"LoadBalancer"}}'

# Use NodePort
kubectl patch svc prometheus-grafana -n monitoring -p '{"spec":{"type":"NodePort"}}'
kubectl get svc prometheus-grafana -n monitoring
```

### **Issue: Inter-Pod Communication Failures**
```
AgentaFlow scheduler cannot reach GPU nodes
```

#### **Network Policy Issues:**
```bash
# Check network policies
kubectl get networkpolicies --all-namespaces

# Temporarily disable to test
kubectl delete networkpolicy --all -n agentaflow-system
```

#### **DNS Resolution:**
```bash
# Test DNS from scheduler pod
kubectl exec -n agentaflow-system deployment/agentaflow-scheduler -- nslookup kubernetes.default.svc.cluster.local

# Check CoreDNS
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl logs -n kube-system -l k8s-app=kube-dns
```

---

## 🚨 Emergency Recovery

### **Issue: Complete Demo Failure**
```
Multiple components not working, need to restart
```

#### **Quick Recovery Script:**
```bash
#!/bin/bash
# emergency-recovery.sh

echo "=== Emergency Demo Recovery ==="

# Step 1: Clean up failed deployments
kubectl delete pods --all -n agentaflow-system --force --grace-period=0
kubectl delete pods --all -n gpu-operator --force --grace-period=0

# Step 2: Restart core components
kubectl rollout restart daemonset/nvidia-device-plugin-daemonset -n gpu-operator
kubectl rollout restart deployment/agentaflow-scheduler -n agentaflow-system
kubectl rollout restart deployment/prometheus-server -n monitoring

# Step 3: Wait for components to be ready
kubectl wait --for=condition=available --timeout=300s deployment/agentaflow-scheduler -n agentaflow-system
kubectl wait --for=condition=ready --timeout=300s pod -l app=nvidia-device-plugin-daemonset -n gpu-operator

# Step 4: Redeploy demo workloads
kubectl delete -f examples/k8s/workloads/ --ignore-not-found=true
sleep 30
kubectl apply -f examples/k8s/workloads/

echo "=== Recovery Complete ==="
```

### **Issue: Need to Start Over**
```
Complete cluster recreation required
```

#### **Full Reset Procedure:**
```bash
# 1. Delete EKS cluster
eksctl delete cluster agentaflow-demo --region us-west-2 --wait

# 2. Clean up any remaining resources
aws ec2 describe-security-groups --filters Name=group-name,Values=eksctl-agentaflow-demo-* \
  --query 'SecurityGroups[*].GroupId' --output text | \
  xargs -I {} aws ec2 delete-security-group --group-id {}

# 3. Wait 5 minutes for cleanup
sleep 300

# 4. Start fresh with the demo setup
# Re-run the setup commands from demo.md
```

---

## 📞 Getting Additional Help

### **Gather Debug Information**
```bash
#!/bin/bash
# gather-debug-info.sh

echo "=== AgentaFlow Demo Debug Information ==="

# System information
echo "## System Info"
kubectl version
eksctl version
helm version
aws --version

# Cluster information  
echo "## Cluster Info"
kubectl get nodes -o wide
kubectl get pods --all-namespaces | grep -E "(agentaflow|gpu-operator|monitoring)"

# Resource usage
echo "## Resource Usage"
kubectl top nodes
kubectl top pods --all-namespaces

# Recent events
echo "## Recent Events"
kubectl get events --sort-by='.lastTimestamp' | tail -20

# Configuration
echo "## Configuration"
kubectl get configmap -n agentaflow-system -o yaml
kubectl describe nodes -l nvidia.com/gpu.present=true

echo "=== Debug Info Collection Complete ==="
```

### **Community Support**
- **GitHub Issues**: [agentaflow-sro-community/issues](https://github.com/Finoptimize/agentaflow-sro-community/issues)
- **Discussions**: [agentaflow-sro-community/discussions](https://github.com/Finoptimize/agentaflow-sro-community/discussions)
- **Documentation**: Check `README.md` and `DOCUMENTATION.md` for additional guidance

### **Enterprise Support**
For production deployments or enterprise evaluation:
- **Email**: [Contact information]
- **Slack**: [Invite link to community Slack]
- **Enterprise Support**: Dedicated support channels for enterprise evaluations

---

## 📋 Common Error Messages Reference

| Error Message | Section | Quick Fix |
|---------------|---------|-----------|
| `InvalidKeyPair.NotFound` | [SSH Key Issues](#issue-ssh-key-not-found) | Create AWS key pair |
| `AccessDenied when calling CreateCluster` | [AWS Permissions](#issue-insufficient-aws-permissions) | Check IAM policies |
| `no spot instance pools available` | [Spot Instances](#issue-spot-instance-unavailable) | Use multiple instance types |
| `Command 'eksctl' not found` | [Missing Tools](#issue-missing-tools) | Install required tools |
| `CrashLoopBackOff` | [AgentaFlow Issues](#issue-agentaflow-scheduler-not-starting) | Check logs and RBAC |
| `ImagePullBackOff` | [Workload Issues](#issue-demo-workloads-failing-to-start) | Verify image availability |
| `CUDA out of memory` | [GPU Memory](#issue-gpu-memory-allocation-errors) | Reduce resource requests |
| `Connection refused` | [Networking](#issue-cannot-access-grafana-dashboard) | Check port forwarding |
| `No data points` | [Monitoring](#issue-grafana-dashboard-not-loading) | Verify Prometheus connection |
| `$0.00 costs` | [Cost Tracking](#issue-cost-data-not-appearing) | Check AWS permissions |

---

**Still having issues? Don't hesitate to reach out for support! The AgentaFlow community is here to help you get your demo running successfully.** 🚀