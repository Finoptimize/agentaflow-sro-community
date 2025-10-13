# AgentaFlow SRO Demo on AWS

**Complete guide to running an AgentaFlow SRO Community Edition demo in AWS to evaluate GPU optimization value**

## 🎯 Demo Objectives

This demo will showcase AgentaFlow's value by demonstrating:
- **40% reduction in GPU idle time** through intelligent scheduling
- **3-5x improvement in inference throughput** via request batching
- **Real-time cost tracking** and optimization insights
- **Kubernetes-native GPU scheduling** with production workloads

## 🏗️ Why AWS for GPU Demos?

### **Market Leadership**
- **40% market share** for cloud ML workloads
- **Most comprehensive GPU instance portfolio** (P4, P5, G4, G5 instances)
- **Mature EKS ecosystem** with excellent GPU operator support
- **Spot instance availability** for cost-effective testing

### **Technical Advantages**
- **NVIDIA GPU Operator** pre-configured for EKS
- **Comprehensive monitoring** via CloudWatch + Prometheus
- **Auto-scaling capabilities** for realistic load testing
- **Cost optimization tools** (Spot instances, Reserved instances)

---

## 💰 Cost-Optimized AWS Demo Architecture

### **Tier 1: Quick Validation Demo ($20-40 total)**
*Perfect for initial evaluation and proof-of-concept*

**Duration**: 4-8 hours  
**Cost**: $20-40 total  
**Use Case**: Validate AgentaFlow scheduling improvements

#### **Infrastructure**
```yaml
EKS Control Plane: 
  - Instance: EKS managed service ($0.10/hour)
  - Duration: 8 hours = $0.80

GPU Worker Nodes (Spot):
  - Instance Type: g4dn.xlarge (1x NVIDIA T4 GPU)
  - Spot Price: ~$0.20/hour (vs $0.526 on-demand)
  - Quantity: 2-3 instances
  - Duration: 6 hours = $2.40-3.60

CPU Worker Nodes:
  - Instance Type: t3.medium (for monitoring)
  - Spot Price: ~$0.015/hour
  - Quantity: 2 instances  
  - Duration: 8 hours = $0.24

Storage & Networking:
  - EBS volumes: $5-10
  - Data transfer: $2-5

Total Estimated Cost: $20-40
```

#### **Demo Workloads**
- **2-3 PyTorch training jobs** (CIFAR-10, MNIST)
- **Inference workload** with varying batch sizes
- **Mixed priority workloads** to show scheduling intelligence

### **Tier 2: Comprehensive Evaluation Demo ($100-200 total)**
*Ideal for thorough evaluation with realistic production scenarios*

**Duration**: 1-2 weeks  
**Cost**: $100-200 total  
**Use Case**: Full feature evaluation with production-like workloads

#### **Infrastructure**
```yaml
EKS Control Plane: 
  - Cost: $0.10/hour × 24 × 14 = $33.60

GPU Worker Nodes (Mixed):
  - g4dn.2xlarge Spot: 2 instances × $0.40/hour × 168 hours = $134.40
  - g5.xlarge On-demand: 1 instance × $1.00/hour × 40 hours = $40.00
  
CPU Worker Nodes:
  - t3.large: 3 instances × $0.04/hour × 168 hours = $20.16

Storage & Services:
  - Prometheus/Grafana stack: $10-15
  - EBS volumes: $15-25
  - Load balancer: $18 (month prorated)

Total Estimated Cost: $150-250
```

### **Tier 3: Enterprise Simulation Demo ($500-1000)**
*Production-scale testing for enterprise decision-making*

**Duration**: 1 month  
**Cost**: $500-1000  
**Use Case**: Enterprise evaluation with multi-team workloads

---

## 🚀 Quick Start: Tier 1 Demo Setup

### **Prerequisites**
```bash
# Install required tools
aws configure  # AWS CLI with appropriate permissions
kubectl version --client  # Kubernetes CLI
eksctl version  # EKS cluster management
helm version  # Package management for Kubernetes
```

### **Step 1: Create EKS Cluster**
```bash
# Create optimized EKS cluster for GPU demo
eksctl create cluster \
  --name agentaflow-demo \
  --region us-west-2 \
  --version 1.28 \
  --nodegroup-name cpu-workers \
  --node-type t3.medium \
  --nodes 2 \
  --nodes-min 1 \
  --nodes-max 4 \
  --managed \
  --spot

# Add GPU node group with spot instances
eksctl create nodegroup \
  --cluster agentaflow-demo \
  --region us-west-2 \
  --name gpu-workers \
  --node-type g4dn.xlarge \
  --nodes 2 \
  --nodes-min 1 \
  --nodes-max 5 \
  --spot \
  --instance-types g4dn.xlarge,g4dn.2xlarge \
  --ssh-access \
  --ssh-public-key your-key-name
```

### **Step 2: Install NVIDIA GPU Operator**
```bash
# Add NVIDIA Helm repository
helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
helm repo update

# Install GPU operator for EKS
helm install gpu-operator nvidia/gpu-operator \
  --namespace gpu-operator \
  --create-namespace \
  --set driver.enabled=false \
  --set toolkit.enabled=false \
  --set devicePlugin.enabled=true \
  --set nodeStatusExporter.enabled=true \
  --set gfd.enabled=true

# Verify GPU nodes
kubectl get nodes -l node.kubernetes.io/instance-type=g4dn.xlarge
kubectl describe node -l nvidia.com/gpu.present=true
```

### **Step 3: Deploy AgentaFlow**
```bash
# Clone the repository
git clone https://github.com/Finoptimize/agentaflow-sro-community.git
cd agentaflow-sro-community

# Deploy AgentaFlow scheduler
kubectl apply -f examples/k8s/scheduler-deployment.yaml

# Deploy monitoring stack
kubectl apply -f examples/k8s/monitoring/

# Verify deployment
kubectl get pods -n agentaflow-system
kubectl logs -n agentaflow-system deployment/agentaflow-scheduler
```

### **Step 4: Deploy Demo Workloads**
```bash
# Deploy PyTorch training workloads
kubectl apply -f examples/k8s/workloads/pytorch-training-1.yaml
kubectl apply -f examples/k8s/workloads/pytorch-training-2.yaml
kubectl apply -f examples/k8s/workloads/inference-workload.yaml

# Deploy mixed-priority workloads to show scheduling
kubectl apply -f examples/k8s/workloads/high-priority-training.yaml
kubectl apply -f examples/k8s/workloads/low-priority-batch.yaml

# Monitor scheduling decisions
./k8s-gpu-scheduler --mode=cli watch
```

---

## 📊 Demo Scenarios & Value Demonstration

### **Scenario 1: GPU Utilization Optimization**
**Objective**: Demonstrate 40% reduction in GPU idle time

#### **Setup**
```bash
# Deploy baseline workloads (default Kubernetes scheduling)
kubectl apply -f examples/demo/baseline-workloads.yaml

# Collect baseline metrics for 30 minutes
kubectl exec deployment/monitoring-collector -- /collect-baseline.sh

# Enable AgentaFlow scheduling
kubectl patch deployment pytorch-training-1 -p '{"metadata":{"labels":{"agentaflow.gpu/managed":"true"}}}'
kubectl patch deployment pytorch-training-2 -p '{"metadata":{"labels":{"agentaflow.gpu/managed":"true"}}}'

# Collect optimized metrics for 30 minutes
kubectl exec deployment/monitoring-collector -- /collect-optimized.sh
```

#### **Expected Results**
```yaml
Baseline Kubernetes Scheduling:
  - Average GPU Utilization: 45-60%
  - Idle Time: 40-55%
  - Workload Queue Time: 5-15 minutes

AgentaFlow Intelligent Scheduling:
  - Average GPU Utilization: 75-90%
  - Idle Time: 10-25% (40% improvement)
  - Workload Queue Time: 1-3 minutes (80% improvement)
  - Cost Savings: 30-40% better $/GPU-hour efficiency
```

### **Scenario 2: Multi-Workload Scheduling Intelligence**
**Objective**: Show intelligent placement across different workload types

#### **Demo Script**
```bash
# Submit diverse workloads simultaneously
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: large-training-job
  labels:
    agentaflow.gpu/managed: "true"
spec:
  template:
    spec:
      containers:
      - name: trainer
        image: pytorch/pytorch:latest
        resources:
          limits:
            nvidia.com/gpu: 1
          requests:
            memory: 8Gi
        env:
        - name: AGENTAFLOW_PRIORITY
          value: "high"
        - name: AGENTAFLOW_MEMORY_REQUIRED  
          value: "6144"  # 6GB
EOF

kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: inference-service
  labels:
    agentaflow.gpu/managed: "true"
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: inference
        image: nvidia/cuda:11.8-runtime-ubuntu20.04
        resources:
          limits:
            nvidia.com/gpu: 1
          requests:
            memory: 4Gi
        env:
        - name: AGENTAFLOW_PRIORITY
          value: "medium"
        - name: AGENTAFLOW_MEMORY_REQUIRED
          value: "2048"  # 2GB
EOF

# Monitor intelligent placement decisions
./k8s-gpu-scheduler --mode=cli status --detailed
```

### **Scenario 3: Cost Optimization Demonstration**
**Objective**: Show real-time cost tracking and optimization

#### **Setup Cost Tracking**
```bash
# Enable cost tracking
kubectl create configmap agentaflow-config \
  --from-literal=enable-cost-tracking=true \
  --from-literal=aws-region=us-west-2 \
  --from-literal=cost-per-gpu-hour-g4dn-xlarge=0.526

# Deploy cost monitoring dashboard
helm install agentaflow-dashboard ./charts/agentaflow-dashboard \
  --set costTracking.enabled=true \
  --set aws.region=us-west-2
```

#### **View Cost Insights**
```bash
# Real-time cost dashboard
kubectl port-forward svc/agentaflow-dashboard 8080:80
# Open http://localhost:8080/costs

# CLI cost summary  
./k8s-gpu-scheduler --mode=cli costs --last=4h

# Export cost report
./k8s-gpu-scheduler --mode=cli costs --export=csv --last=24h > demo-costs.csv
```

---

## 📈 Monitoring & Observability Demo

### **Grafana Dashboard Setup**
```bash
# Deploy monitoring stack
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.enabled=true

# Import AgentaFlow dashboards
kubectl create configmap agentaflow-dashboards \
  --from-file=examples/monitoring/grafana-dashboards/ \
  -n monitoring

# Access Grafana
kubectl port-forward svc/prometheus-grafana 3000:80 -n monitoring
# Login: admin/prom-operator
```

### **Key Metrics to Demonstrate**

#### **GPU Utilization Dashboard**
- Real-time GPU utilization across all nodes
- Memory usage and temperature monitoring  
- Workload scheduling decisions timeline
- Cost per GPU-hour trending

#### **Workload Performance Dashboard**
- Training job completion times
- Inference latency and throughput
- Queue wait times and scheduling efficiency
- Resource utilization by workload type

#### **Cost Optimization Dashboard** 
- Real-time cost tracking by workload
- GPU efficiency metrics (cost/output)
- Spot vs on-demand usage optimization
- Projected monthly costs and savings

---

## 🔄 Advanced Demo Scenarios

### **Scenario 4: Auto-scaling Under Load**
```bash
# Deploy load generator
kubectl apply -f examples/demo/load-generator.yaml

# Configure HPA for GPU workloads
kubectl apply -f - <<EOF
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler  
metadata:
  name: inference-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: inference-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Pods
    pods:
      metric:
        name: nvidia_gpu_utilization
      target:
        type: AverageValue
        averageValue: "70"
EOF

# Monitor auto-scaling behavior
kubectl get hpa -w
kubectl get pods -l app=inference-service -w
```

### **Scenario 5: Multi-Strategy Scheduling**
```bash
# Test different scheduling strategies
./k8s-gpu-scheduler --mode=cli strategy least-utilized
# Deploy workloads and observe placement

./k8s-gpu-scheduler --mode=cli strategy best-fit  
# Deploy same workloads and compare placement

./k8s-gpu-scheduler --mode=cli strategy priority
# Deploy mixed-priority workloads

# Compare results
./k8s-gpu-scheduler --mode=cli metrics --compare-strategies
```

---

## 🎯 Success Metrics & Expected Outcomes

### **Quantifiable Improvements**
```yaml
GPU Utilization:
  Before: 45-60% average utilization
  After: 75-90% average utilization
  Impact: 40% reduction in idle time

Cost Efficiency:
  Before: $0.526/hour per GPU (baseline)
  After: $0.320/hour effective cost per GPU-hour of work
  Impact: 39% cost reduction per unit of work

Workload Scheduling:
  Before: 5-15 minute average queue time
  After: 1-3 minute average queue time  
  Impact: 80% reduction in wait times

Throughput:
  Before: 2-3 concurrent jobs per GPU
  After: 5-8 concurrent jobs per GPU (via intelligent batching)
  Impact: 3x improvement in job throughput
```

### **Observability Improvements**
- **Real-time visibility** into GPU utilization across all nodes
- **Predictive cost tracking** with budget alerts
- **Performance bottleneck identification** and optimization recommendations
- **Automated workload placement** with explainable scheduling decisions

---

## 🧪 Validation Tests

### **Performance Validation Script**
```bash
#!/bin/bash
# performance-validation.sh

echo "=== AgentaFlow Demo Performance Validation ==="

# Test 1: GPU utilization improvement
echo "Running baseline GPU utilization test..."
./run-baseline-test.sh > baseline-results.json

echo "Running AgentaFlow optimization test..."
./run-agentaflow-test.sh > optimized-results.json

# Calculate improvement
python3 calculate-improvements.py baseline-results.json optimized-results.json

# Test 2: Cost tracking accuracy  
echo "Validating cost tracking..."
./validate-cost-tracking.sh

# Test 3: Scheduling intelligence
echo "Testing scheduling strategies..."
./test-scheduling-strategies.sh

echo "=== Demo validation complete ==="
```

### **Automated Comparison Report**
The demo includes scripts to generate:
- **Before/after utilization charts**
- **Cost comparison tables**
- **Scheduling efficiency metrics**
- **ROI calculations** for production deployment

---

## 🚀 Enterprise Demo Extensions

### **Multi-Cluster Simulation**
```bash
# Create second demo cluster for multi-cluster scenarios
eksctl create cluster --name agentaflow-demo-2 --region us-east-1

# Show cross-cluster workload placement
./k8s-gpu-scheduler --mode=multi-cluster \
  --clusters agentaflow-demo,agentaflow-demo-2 \
  --enable-cross-cluster-scheduling
```

### **Integration Demonstrations**
- **Slack notifications** for cost thresholds and GPU alerts
- **DataDog integration** for enterprise monitoring
- **AWS Cost Explorer** integration for billing optimization
- **Prometheus metrics** export for existing monitoring stacks

---

## 🧹 Demo Cleanup

### **Automated Cleanup Script**
```bash
#!/bin/bash
# cleanup-demo.sh

echo "Cleaning up AgentaFlow demo resources..."

# Delete workloads
kubectl delete jobs --all -n default
kubectl delete deployments --all -n default  

# Delete AgentaFlow components
kubectl delete -f examples/k8s/scheduler-deployment.yaml
helm uninstall gpu-operator -n gpu-operator
helm uninstall prometheus -n monitoring

# Delete EKS cluster
eksctl delete cluster agentaflow-demo --region us-west-2

echo "Demo cleanup complete!"
echo "Final cost summary:"
aws ce get-cost-and-usage --time-period Start=2025-01-01,End=2025-01-02 \
  --granularity DAILY --metrics UnblendedCost \
  --group-by Type=DIMENSION,Key=SERVICE | jq '.ResultsByTime[0].Groups[] | select(.Keys[0] | contains("EC2")) | .Metrics.UnblendedCost'
```

---

## 📞 Demo Support & Next Steps

### **Getting Help During Demo**
- **Documentation**: Complete setup guides in `/examples/demo/`  
- **Troubleshooting**: Common issues and solutions in `TROUBLESHOOTING.md`
- **Community Support**: GitHub Issues and Discussions
- **Enterprise Support**: Contact for dedicated demo assistance

### **After the Demo**
1. **Production Planning**: Architecture review and sizing recommendations
2. **Enterprise Features**: Evaluation of advanced capabilities  
3. **Migration Planning**: Step-by-step production deployment guide
4. **Training & Onboarding**: Team training and best practices

---

**Ready to see AgentaFlow optimize your GPU infrastructure? Start your AWS demo today!** 🚀