#!/bin/bash

# setup-demo.sh - Automated AgentaFlow AWS Demo Setup
# This script creates a complete AWS EKS environment for demonstrating AgentaFlow

set -e

echo "🚀 AgentaFlow AWS Demo Setup Starting..."

# Configuration
CLUSTER_NAME="agentaflow-demo"
REGION="us-west-2"
NODE_TYPE_CPU="t3.medium"
NODE_TYPE_GPU="g4dn.xlarge"

# Color coding for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    command -v aws >/dev/null 2>&1 || { 
        log_error "AWS CLI not installed. Please install it first."
        exit 1
    }
    
    command -v kubectl >/dev/null 2>&1 || { 
        log_error "kubectl not installed. Please install it first."
        exit 1
    }
    
    command -v eksctl >/dev/null 2>&1 || { 
        log_error "eksctl not installed. Please install it first."
        exit 1
    }
    
    command -v helm >/dev/null 2>&1 || { 
        log_error "Helm not installed. Please install it first."
        exit 1
    }
    
    # Check AWS credentials
    aws sts get-caller-identity >/dev/null 2>&1 || {
        log_error "AWS credentials not configured. Run 'aws configure' first."
        exit 1
    }
    
    log_info "✓ All prerequisites met"
}

# Create EKS cluster
create_eks_cluster() {
    log_info "Creating EKS cluster: $CLUSTER_NAME"
    
    # Check if cluster already exists
    if eksctl get cluster --name $CLUSTER_NAME --region $REGION >/dev/null 2>&1; then
        log_warn "Cluster $CLUSTER_NAME already exists. Skipping creation."
        return 0
    fi
    
    log_info "Creating CPU worker node group..."
    eksctl create cluster \
        --name $CLUSTER_NAME \
        --region $REGION \
        --version 1.28 \
        --nodegroup-name cpu-workers \
        --node-type $NODE_TYPE_CPU \
        --nodes 2 \
        --nodes-min 1 \
        --nodes-max 4 \
        --managed \
        --spot
    
    log_info "Adding GPU worker node group..."
    eksctl create nodegroup \
        --cluster $CLUSTER_NAME \
        --region $REGION \
        --name gpu-workers \
        --node-type $NODE_TYPE_GPU \
        --nodes 2 \
        --nodes-min 1 \
        --nodes-max 5 \
        --spot \
        --instance-types $NODE_TYPE_GPU
    
    log_info "✓ EKS cluster created successfully"
}

# Install NVIDIA GPU Operator
install_gpu_operator() {
    log_info "Installing NVIDIA GPU Operator..."
    
    helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
    helm repo update
    
    helm install gpu-operator nvidia/gpu-operator \
        --namespace gpu-operator \
        --create-namespace \
        --set driver.enabled=false \
        --set toolkit.enabled=false \
        --set devicePlugin.enabled=true \
        --set nodeStatusExporter.enabled=true \
        --set gfd.enabled=true \
        --wait \
        --timeout=10m
    
    log_info "✓ GPU Operator installed"
}

# Deploy AgentaFlow
deploy_agentaflow() {
    log_info "Deploying AgentaFlow scheduler..."
    
    # Create namespace
    kubectl create namespace agentaflow-system --dry-run=client -o yaml | kubectl apply -f -
    
    # Build and deploy AgentaFlow scheduler
    kubectl apply -f ../../k8s/scheduler-deployment.yaml
    
    # Wait for deployment to be ready
    kubectl wait --for=condition=available --timeout=300s deployment/agentaflow-scheduler -n agentaflow-system
    
    log_info "✓ AgentaFlow deployed"
}

# Install monitoring stack
install_monitoring() {
    log_info "Installing monitoring stack (Prometheus + Grafana)..."
    
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update
    
    helm install prometheus prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --create-namespace \
        --set grafana.enabled=true \
        --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
        --wait \
        --timeout=10m
    
    log_info "✓ Monitoring stack installed"
}

# Deploy demo workloads
deploy_demo_workloads() {
    log_info "Deploying demo workloads..."
    
    # Apply baseline workloads
    kubectl apply -f ../k8s/workloads/baseline-workloads.yaml
    
    # Apply AgentaFlow managed workloads
    kubectl apply -f ../k8s/workloads/agentaflow-workloads.yaml
    
    log_info "✓ Demo workloads deployed"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    # Check nodes
    log_info "GPU Nodes:"
    kubectl get nodes -l node.kubernetes.io/instance-type=$NODE_TYPE_GPU
    
    # Check GPU resources
    log_info "GPU Resources:"
    kubectl describe nodes -l nvidia.com/gpu.present=true | grep "nvidia.com/gpu"
    
    # Check AgentaFlow
    log_info "AgentaFlow Scheduler:"
    kubectl get pods -n agentaflow-system
    
    # Check monitoring
    log_info "Monitoring Stack:"
    kubectl get pods -n monitoring | grep -E "(prometheus|grafana)"
    
    log_info "✓ Deployment verified"
}

# Print access information
print_access_info() {
    log_info "🎉 Demo setup complete!"
    echo
    echo "Access Information:"
    echo "=================="
    
    # Get Grafana admin password
    GRAFANA_PASSWORD=$(kubectl get secret --namespace monitoring prometheus-grafana -o jsonpath="{.data.admin-password}" | base64 --decode)
    
    echo "📊 Grafana Dashboard:"
    echo "   Port forward: kubectl port-forward svc/prometheus-grafana 3000:80 -n monitoring"
    echo "   URL: http://localhost:3000"
    echo "   Username: admin"
    echo "   Password: $GRAFANA_PASSWORD"
    echo
    
    echo "🔍 AgentaFlow CLI:"
    echo "   Status: ./k8s-gpu-scheduler --mode=cli status"
    echo "   Watch: ./k8s-gpu-scheduler --mode=cli watch"
    echo "   Health: ./k8s-gpu-scheduler --mode=cli health"
    echo
    
    echo "🏷️  Useful Commands:"
    echo "   View GPU nodes: kubectl get nodes -l nvidia.com/gpu.present=true"
    echo "   Check workloads: kubectl get pods --show-labels"
    echo "   Monitor costs: ./k8s-gpu-scheduler --mode=cli costs"
    echo
    
    log_warn "💰 Remember to run ./cleanup-demo.sh when finished to avoid ongoing costs!"
}

# Main execution
main() {
    echo "AgentaFlow AWS Demo Setup"
    echo "========================"
    echo "Cluster: $CLUSTER_NAME"
    echo "Region: $REGION"
    echo "GPU Instance: $NODE_TYPE_GPU"
    echo
    
    read -p "Continue with setup? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Setup cancelled."
        exit 0
    fi
    
    check_prerequisites
    create_eks_cluster
    install_gpu_operator
    deploy_agentaflow
    install_monitoring
    deploy_demo_workloads
    verify_deployment
    print_access_info
}

# Run main function
main "$@"