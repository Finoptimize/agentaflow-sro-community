#!/usr/bin/env python3

"""
cost-calculator.py - AgentaFlow Demo Cost Analysis Tool

This script analyzes AWS costs for the AgentaFlow demo and calculates
ROI metrics based on GPU utilization improvements.
"""

import json
import argparse
from datetime import datetime, timedelta
from typing import Dict, List, Optional
import boto3
from botocore.exceptions import ClientError, NoCredentialsError

class DemoCostCalculator:
    def __init__(self, region: str = "us-west-2"):
        self.region = region
        try:
            self.ce_client = boto3.client('ce', region_name='us-east-1')  # Cost Explorer is in us-east-1
            self.ec2_client = boto3.client('ec2', region_name=region)
        except NoCredentialsError:
            print("❌ AWS credentials not configured. Some features will be unavailable.")
            self.ce_client = None
            self.ec2_client = None

    def get_demo_costs(self, start_date: str, end_date: str) -> Dict:
        """Fetch AWS costs for the demo period"""
        if not self.ce_client:
            return {"error": "AWS credentials not available"}
        
        try:
            response = self.ce_client.get_cost_and_usage(
                TimePeriod={
                    'Start': start_date,
                    'End': end_date
                },
                Granularity='DAILY',
                Metrics=['UnblendedCost'],
                GroupBy=[
                    {'Type': 'DIMENSION', 'Key': 'SERVICE'},
                    {'Type': 'DIMENSION', 'Key': 'INSTANCE_TYPE'}
                ]
            )
            return response
        except ClientError as e:
            return {"error": f"Failed to fetch cost data: {e}"}

    def calculate_gpu_instance_costs(self, instance_types: List[str], hours: float) -> Dict[str, float]:
        """Calculate costs for different GPU instance types"""
        # AWS EC2 pricing (approximate, varies by region)
        pricing = {
            'g4dn.xlarge': 0.526,    # 1x T4 GPU
            'g4dn.2xlarge': 0.752,   # 1x T4 GPU, more CPU/RAM
            'g4dn.4xlarge': 1.204,   # 1x T4 GPU, even more resources
            'g5.xlarge': 1.006,      # 1x A10G GPU
            'g5.2xlarge': 1.212,     # 1x A10G GPU
            'p3.2xlarge': 3.06,      # 1x V100 GPU
            'p4d.24xlarge': 32.77,   # 8x A100 GPUs
        }
        
        costs = {}
        for instance_type in instance_types:
            if instance_type in pricing:
                costs[instance_type] = pricing[instance_type] * hours
            else:
                costs[instance_type] = f"Unknown pricing for {instance_type}"
        
        return costs

    def calculate_utilization_improvement(self, baseline_util: float, optimized_util: float) -> Dict:
        """Calculate the improvement metrics from GPU optimization"""
        improvement_percent = ((optimized_util - baseline_util) / baseline_util) * 100
        cost_efficiency = optimized_util / baseline_util
        idle_time_reduction = (1 - (1 - optimized_util) / (1 - baseline_util)) * 100
        
        return {
            'utilization_improvement_percent': improvement_percent,
            'cost_efficiency_multiplier': cost_efficiency,
            'idle_time_reduction_percent': idle_time_reduction,
            'effective_cost_per_work_unit': 1 / cost_efficiency
        }

    def generate_roi_analysis(self, monthly_gpu_spend: float, utilization_improvement: float) -> Dict:
        """Generate ROI analysis for AgentaFlow deployment"""
        # Calculate potential monthly savings
        current_efficiency = 0.55  # Typical baseline GPU utilization
        improved_efficiency = current_efficiency * (1 + utilization_improvement / 100)
        
        # Cost savings from better utilization
        cost_reduction_percent = (improved_efficiency - current_efficiency) / current_efficiency * 100
        monthly_savings = monthly_gpu_spend * (cost_reduction_percent / 100)
        annual_savings = monthly_savings * 12
        
        # AgentaFlow operational costs (estimated)
        agentaflow_monthly_cost = 50  # Monitoring infrastructure
        net_monthly_savings = monthly_savings - agentaflow_monthly_cost
        roi_months = agentaflow_monthly_cost / monthly_savings if monthly_savings > 0 else float('inf')
        
        return {
            'current_utilization': current_efficiency * 100,
            'improved_utilization': improved_efficiency * 100,
            'cost_reduction_percent': cost_reduction_percent,
            'monthly_savings': monthly_savings,
            'annual_savings': annual_savings,
            'agentaflow_monthly_cost': agentaflow_monthly_cost,
            'net_monthly_savings': net_monthly_savings,
            'payback_period_months': roi_months,
            'annual_roi_percent': (annual_savings / (agentaflow_monthly_cost * 12)) * 100 if monthly_savings > 0 else 0
        }

    def generate_demo_report(self, demo_config: Dict) -> Dict:
        """Generate comprehensive demo cost and performance report"""
        report = {
            'demo_info': demo_config,
            'timestamp': datetime.utcnow().isoformat(),
            'calculations': {}
        }
        
        # Calculate demo infrastructure costs
        gpu_hours = demo_config.get('duration_hours', 8)
        instance_type = demo_config.get('instance_type', 'g4dn.xlarge')
        node_count = demo_config.get('node_count', 2)
        
        instance_costs = self.calculate_gpu_instance_costs([instance_type], gpu_hours)
        total_gpu_cost = instance_costs.get(instance_type, 0) * node_count
        
        # Add EKS and other costs
        eks_cost = 0.10 * gpu_hours  # EKS control plane
        monitoring_cost = 5  # Prometheus/Grafana
        networking_cost = 2  # Data transfer
        storage_cost = 3     # EBS volumes
        
        total_demo_cost = total_gpu_cost + eks_cost + monitoring_cost + networking_cost + storage_cost
        
        report['calculations'] = {
            'infrastructure_costs': {
                'gpu_instances': total_gpu_cost,
                'eks_control_plane': eks_cost,
                'monitoring_stack': monitoring_cost,
                'networking': networking_cost,
                'storage': storage_cost,
                'total': total_demo_cost
            }
        }
        
        # Calculate performance improvements
        baseline_util = demo_config.get('baseline_utilization', 55)
        optimized_util = demo_config.get('optimized_utilization', 85)
        
        improvement_metrics = self.calculate_utilization_improvement(
            baseline_util / 100, optimized_util / 100
        )
        
        report['calculations']['performance_improvements'] = improvement_metrics
        
        # ROI analysis for production deployment
        monthly_gpu_spend = demo_config.get('monthly_gpu_spend', 10000)
        roi_analysis = self.generate_roi_analysis(
            monthly_gpu_spend, 
            improvement_metrics['utilization_improvement_percent']
        )
        
        report['calculations']['roi_analysis'] = roi_analysis
        
        return report

def main():
    parser = argparse.ArgumentParser(description='AgentaFlow Demo Cost Calculator')
    parser.add_argument('--config', type=str, help='Demo configuration JSON file')
    parser.add_argument('--monthly-spend', type=float, default=10000, 
                       help='Monthly GPU spend for ROI calculation')
    parser.add_argument('--output', type=str, default='demo-analysis.json',
                       help='Output file for analysis report')
    parser.add_argument('--region', type=str, default='us-west-2',
                       help='AWS region for cost analysis')
    
    args = parser.parse_args()
    
    calculator = DemoCostCalculator(region=args.region)
    
    # Default demo configuration
    demo_config = {
        'cluster_name': 'agentaflow-demo',
        'region': args.region,
        'duration_hours': 8,
        'instance_type': 'g4dn.xlarge',
        'node_count': 2,
        'baseline_utilization': 55,  # Typical Kubernetes scheduling
        'optimized_utilization': 85, # With AgentaFlow
        'monthly_gpu_spend': args.monthly_spend
    }
    
    # Load custom configuration if provided
    if args.config:
        try:
            with open(args.config, 'r') as f:
                custom_config = json.load(f)
                demo_config.update(custom_config)
        except FileNotFoundError:
            print(f"⚠️  Config file {args.config} not found, using defaults")
    
    # Generate comprehensive report
    print("📊 Generating AgentaFlow demo cost analysis...")
    report = calculator.generate_demo_report(demo_config)
    
    # Save report to file
    with open(args.output, 'w') as f:
        json.dump(report, f, indent=2)
    
    # Print summary to console
    print("\n" + "="*60)
    print("🎯 AGENTAFLOW DEMO COST ANALYSIS SUMMARY")
    print("="*60)
    
    costs = report['calculations']['infrastructure_costs']
    print(f"💰 Demo Infrastructure Costs:")
    print(f"   GPU Instances ({demo_config['node_count']}x {demo_config['instance_type']}): ${costs['gpu_instances']:.2f}")
    print(f"   EKS Control Plane: ${costs['eks_control_plane']:.2f}")
    print(f"   Monitoring & Other: ${costs['monitoring_stack'] + costs['networking'] + costs['storage']:.2f}")
    print(f"   Total Demo Cost: ${costs['total']:.2f}")
    
    perf = report['calculations']['performance_improvements']
    print(f"\n📈 Performance Improvements:")
    print(f"   Utilization Improvement: {perf['utilization_improvement_percent']:.1f}%")
    print(f"   Cost Efficiency Gain: {perf['cost_efficiency_multiplier']:.2f}x")
    print(f"   Idle Time Reduction: {perf['idle_time_reduction_percent']:.1f}%")
    
    roi = report['calculations']['roi_analysis']
    print(f"\n🚀 Production ROI Analysis:")
    print(f"   Monthly GPU Spend: ${demo_config['monthly_gpu_spend']:,.0f}")
    print(f"   Projected Monthly Savings: ${roi['monthly_savings']:,.0f}")
    print(f"   Annual Savings: ${roi['annual_savings']:,.0f}")
    print(f"   Payback Period: {roi['payback_period_months']:.1f} months")
    print(f"   Annual ROI: {roi['annual_roi_percent']:.0f}%")
    
    print(f"\n📋 Detailed report saved to: {args.output}")
    print("\n🎉 Analysis complete! Use this data to evaluate AgentaFlow's value proposition.")

if __name__ == "__main__":
    main()