# AgentaFlow AWS Demo Resources

This directory contains all the resources needed to run the AgentaFlow AWS demo.

## 📁 Directory Structure

```
demo/
├── README.md                    # This file
├── scripts/
│   ├── setup-demo.sh           # Automated demo setup
│   ├── cleanup-demo.sh         # Resource cleanup
│   ├── performance-validation.sh # Validation testing
│   └── cost-calculator.py      # Cost analysis
├── terraform/
│   ├── main.tf                 # Infrastructure as code
│   ├── variables.tf            # Configuration variables
│   └── outputs.tf              # Resource outputs
├── k8s/
│   ├── baseline-workloads.yaml # Standard Kubernetes scheduling
│   ├── agentaflow-workloads.yaml # AgentaFlow managed workloads
│   ├── monitoring/             # Prometheus & Grafana configs
│   └── load-testing/           # Load generation workloads
└── results/
    ├── sample-baseline.json    # Example baseline results
    └── sample-optimized.json   # Example optimized results
```

## 🚀 Quick Start

1. **Prerequisites**: Ensure AWS CLI, kubectl, eksctl, and Terraform are installed
2. **Setup**: Run `./scripts/setup-demo.sh` to create infrastructure
3. **Demo**: Follow the scenarios in `/demo.md`  
4. **Cleanup**: Run `./scripts/cleanup-demo.sh` when finished

## 📊 Expected Results

The demo typically shows:
- **40% improvement** in GPU utilization
- **30-50% cost reduction** per unit of work
- **80% faster** workload scheduling
- **Real-time visibility** into GPU resource usage

## 🆘 Troubleshooting

Common issues and solutions:

### EKS Cluster Creation Fails
```bash
# Check AWS permissions
aws sts get-caller-identity
# Ensure you have EKS, EC2, and IAM permissions
```

### GPU Nodes Not Joining Cluster
```bash
# Verify GPU operator installation
kubectl get pods -n gpu-operator
# Check node labels
kubectl get nodes --show-labels | grep gpu
```

### Workloads Not Scheduling
```bash
# Check AgentaFlow scheduler logs
kubectl logs -n agentaflow-system deployment/agentaflow-scheduler
# Verify GPU resources
kubectl describe node -l nvidia.com/gpu.present=true
```

## 💰 Cost Estimation

| Demo Tier | Duration | Estimated Cost | Use Case |
|-----------|----------|----------------|----------|
| Tier 1 | 4-8 hours | $20-40 | Quick validation |
| Tier 2 | 1-2 weeks | $100-200 | Comprehensive eval |
| Tier 3 | 1 month | $500-1000 | Enterprise testing |

## 📞 Support

For demo support:
- **GitHub Issues**: Technical problems and questions
- **Documentation**: Complete guides in parent directory
- **Enterprise Support**: Contact for dedicated assistance