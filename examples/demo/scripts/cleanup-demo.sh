#!/bin/bash

# cleanup-demo.sh - Clean up AgentaFlow AWS Demo Resources
# This script removes all resources created by the demo

set -e

# Configuration
CLUSTER_NAME="agentaflow-demo"
REGION="us-west-2"

# Color coding
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Generate final cost report
generate_cost_report() {
    log_info "Generating final cost report..."
    
    # Get today's date for cost analysis
    TODAY=$(date +%Y-%m-%d)
    YESTERDAY=$(date -d "yesterday" +%Y-%m-%d)
    
    log_info "Fetching AWS costs for demo period..."
    
    # Create cost report
    cat << EOF > demo-cost-report.json
{
    "report_date": "$TODAY",
    "cluster_name": "$CLUSTER_NAME",
    "region": "$REGION",
    "period": "$YESTERDAY to $TODAY"
}
EOF
    
    # Try to get actual AWS costs
    if command -v aws >/dev/null 2>&1; then
        aws ce get-cost-and-usage \
            --time-period Start=$YESTERDAY,End=$TODAY \
            --granularity DAILY \
            --metrics UnblendedCost \
            --group-by Type=DIMENSION,Key=SERVICE \
            --output json > aws-cost-raw.json 2>/dev/null || log_warn "Could not fetch AWS cost data"
    fi
    
    log_info "✓ Cost report saved to demo-cost-report.json"
}

# Clean up Kubernetes resources
cleanup_k8s_resources() {
    log_info "Cleaning up Kubernetes resources..."
    
    # Delete demo workloads
    log_info "Removing demo workloads..."
    kubectl delete jobs --all -n default --timeout=60s || true
    kubectl delete deployments --all -n default --timeout=60s || true
    kubectl delete pods --all -n default --timeout=60s || true
    
    # Delete AgentaFlow components
    log_info "Removing AgentaFlow components..."
    kubectl delete -f ../../k8s/scheduler-deployment.yaml --timeout=60s || true
    kubectl delete namespace agentaflow-system --timeout=60s || true
    
    # Uninstall Helm releases
    log_info "Removing Helm releases..."
    helm uninstall gpu-operator -n gpu-operator --timeout=300s || true
    helm uninstall prometheus -n monitoring --timeout=300s || true
    
    # Delete namespaces
    kubectl delete namespace gpu-operator --timeout=60s || true
    kubectl delete namespace monitoring --timeout=60s || true
    
    log_info "✓ Kubernetes resources cleaned up"
}

# Delete EKS cluster
delete_eks_cluster() {
    log_info "Deleting EKS cluster: $CLUSTER_NAME"
    
    # Check if cluster exists
    if ! eksctl get cluster --name $CLUSTER_NAME --region $REGION >/dev/null 2>&1; then
        log_warn "Cluster $CLUSTER_NAME does not exist. Skipping deletion."
        return 0
    fi
    
    # Delete the entire cluster
    log_warn "This will delete the entire EKS cluster and all associated resources."
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cluster deletion cancelled."
        return 0
    fi
    
    eksctl delete cluster --name $CLUSTER_NAME --region $REGION --wait
    
    log_info "✓ EKS cluster deleted"
}

# Clean up additional AWS resources
cleanup_aws_resources() {
    log_info "Checking for additional AWS resources to clean up..."
    
    # Clean up any leftover security groups
    log_info "Checking for leftover security groups..."
    aws ec2 describe-security-groups \
        --region $REGION \
        --filters "Name=group-name,Values=*$CLUSTER_NAME*" \
        --query 'SecurityGroups[].GroupId' \
        --output text 2>/dev/null | while read -r sg_id; do
        if [ ! -z "$sg_id" ] && [ "$sg_id" != "None" ]; then
            log_warn "Found leftover security group: $sg_id"
            echo "Run manually: aws ec2 delete-security-group --group-id $sg_id --region $REGION"
        fi
    done
    
    # Clean up any leftover volumes
    log_info "Checking for leftover EBS volumes..."
    aws ec2 describe-volumes \
        --region $REGION \
        --filters "Name=tag:kubernetes.io/cluster/$CLUSTER_NAME,Values=owned" \
        --query 'Volumes[?State==`available`].VolumeId' \
        --output text 2>/dev/null | while read -r vol_id; do
        if [ ! -z "$vol_id" ] && [ "$vol_id" != "None" ]; then
            log_warn "Found leftover volume: $vol_id"
            echo "Run manually: aws ec2 delete-volume --volume-id $vol_id --region $REGION"
        fi
    done
    
    log_info "✓ AWS resource check complete"
}

# Verify cleanup
verify_cleanup() {
    log_info "Verifying cleanup..."
    
    # Check if cluster still exists
    if eksctl get cluster --name $CLUSTER_NAME --region $REGION >/dev/null 2>&1; then
        log_error "Cluster $CLUSTER_NAME still exists!"
        return 1
    fi
    
    log_info "✓ Cleanup verified"
}

# Print summary
print_summary() {
    log_info "🧹 Demo cleanup complete!"
    echo
    echo "Summary:"
    echo "========"
    echo "✓ Kubernetes workloads removed"
    echo "✓ AgentaFlow components uninstalled"
    echo "✓ Monitoring stack removed"
    echo "✓ EKS cluster deleted"
    echo "✓ Additional AWS resources checked"
    echo
    
    if [ -f "demo-cost-report.json" ]; then
        log_info "📊 Cost report saved: demo-cost-report.json"
    fi
    
    echo "🔍 Manual verification steps:"
    echo "   1. Check AWS console for any leftover resources"
    echo "   2. Review AWS billing for demo period charges"
    echo "   3. Verify no ongoing charges from demo resources"
    echo
    
    log_warn "💡 Tip: Check your AWS billing dashboard to confirm all demo resources are removed"
}

# Main execution
main() {
    echo "AgentaFlow AWS Demo Cleanup"
    echo "==========================="
    echo "Cluster: $CLUSTER_NAME"
    echo "Region: $REGION"
    echo
    
    log_warn "This will DELETE all demo resources including:"
    log_warn "- EKS cluster and worker nodes"
    log_warn "- All deployed workloads and data"
    log_warn "- Monitoring stack and dashboards"
    echo
    
    read -p "Are you sure you want to proceed? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleanup cancelled."
        exit 0
    fi
    
    generate_cost_report
    cleanup_k8s_resources
    delete_eks_cluster
    cleanup_aws_resources
    verify_cleanup
    print_summary
}

# Run main function
main "$@"