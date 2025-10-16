#!/usr/bin/env python3
"""
CI/CD Build Performance Analytics

Analyzes GitHub Actions workflow runs to provide insights on:
- Build duration trends
- Cache hit rates
- Failure patterns
- Resource utilization
- Cost optimization opportunities
"""

import os
import sys
import json
import requests
from datetime import datetime, timedelta
from collections import defaultdict
from typing import Dict, List, Any

# Configuration
GITHUB_API = "https://api.github.com"
REPO_OWNER = os.getenv("GITHUB_REPOSITORY_OWNER", "Finoptimize")
REPO_NAME = os.getenv("GITHUB_REPOSITORY_NAME", "agentaflow-sro-community")
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN", "")

def get_workflow_runs(workflow_name: str, days: int = 30) -> List[Dict]:
    """Fetch workflow runs from GitHub API"""
    headers = {
        "Authorization": f"token {GITHUB_TOKEN}",
        "Accept": "application/vnd.github.v3+json"
    }
    
    url = f"{GITHUB_API}/repos/{REPO_OWNER}/{REPO_NAME}/actions/workflows/{workflow_name}/runs"
    params = {
        "per_page": 100,
        "created": f">={(datetime.now() - timedelta(days=days)).isoformat()}"
    }
    
    response = requests.get(url, headers=headers, params=params)
    response.raise_for_status()
    
    return response.json().get("workflow_runs", [])

def analyze_build_performance(runs: List[Dict]) -> Dict[str, Any]:
    """Analyze build performance metrics"""
    total_runs = len(runs)
    successful_runs = [r for r in runs if r["conclusion"] == "success"]
    failed_runs = [r for r in runs if r["conclusion"] == "failure"]
    
    # Calculate durations
    durations = []
    for run in successful_runs:
        if run["created_at"] and run["updated_at"]:
            created = datetime.fromisoformat(run["created_at"].replace("Z", "+00:00"))
            updated = datetime.fromisoformat(run["updated_at"].replace("Z", "+00:00"))
            duration = (updated - created).total_seconds() / 60  # minutes
            durations.append(duration)
    
    avg_duration = sum(durations) / len(durations) if durations else 0
    min_duration = min(durations) if durations else 0
    max_duration = max(durations) if durations else 0
    
    # Success rate
    success_rate = (len(successful_runs) / total_runs * 100) if total_runs > 0 else 0
    
    # Failure analysis
    failure_reasons = defaultdict(int)
    for run in failed_runs:
        # Extract failure reason from conclusion
        failure_reasons[run.get("conclusion", "unknown")] += 1
    
    return {
        "total_runs": total_runs,
        "successful_runs": len(successful_runs),
        "failed_runs": len(failed_runs),
        "success_rate": round(success_rate, 2),
        "avg_duration_minutes": round(avg_duration, 2),
        "min_duration_minutes": round(min_duration, 2),
        "max_duration_minutes": round(max_duration, 2),
        "failure_reasons": dict(failure_reasons)
    }

def calculate_cache_efficiency(runs: List[Dict]) -> Dict[str, Any]:
    """Estimate cache hit rate from build times"""
    # Builds with cache hits should be significantly faster
    durations = []
    for run in runs:
        if run["conclusion"] == "success" and run["created_at"] and run["updated_at"]:
            created = datetime.fromisoformat(run["created_at"].replace("Z", "+00:00"))
            updated = datetime.fromisoformat(run["updated_at"].replace("Z", "+00:00"))
            duration = (updated - created).total_seconds() / 60
            durations.append(duration)
    
    if not durations:
        return {"cache_hit_rate_estimate": 0, "note": "No data available"}
    
    avg_duration = sum(durations) / len(durations)
    fast_builds = [d for d in durations if d < avg_duration * 0.8]  # 20% faster
    cache_hit_rate = (len(fast_builds) / len(durations) * 100) if durations else 0
    
    return {
        "cache_hit_rate_estimate": round(cache_hit_rate, 2),
        "fast_builds": len(fast_builds),
        "total_builds": len(durations),
        "note": "Estimated based on build duration variance"
    }

def generate_recommendations(metrics: Dict[str, Any]) -> List[str]:
    """Generate optimization recommendations"""
    recommendations = []
    
    # Success rate recommendations
    if metrics["success_rate"] < 90:
        recommendations.append(
            f"⚠️  Success rate is {metrics['success_rate']}% - investigate common failure patterns"
        )
    
    # Duration recommendations
    if metrics["avg_duration_minutes"] > 15:
        recommendations.append(
            f"⚠️  Average build time is {metrics['avg_duration_minutes']} minutes - consider optimization"
        )
    
    # Cache recommendations
    cache_metrics = calculate_cache_efficiency([])
    if cache_metrics.get("cache_hit_rate_estimate", 0) < 80:
        recommendations.append(
            "⚠️  Low cache hit rate detected - review caching strategy"
        )
    
    if not recommendations:
        recommendations.append("✅ All metrics within acceptable ranges")
    
    return recommendations

def main():
    """Main execution"""
    print("=" * 70)
    print("AgentaFlow CI/CD Performance Analytics")
    print("=" * 70)
    print()
    
    if not GITHUB_TOKEN:
        print("⚠️  Warning: GITHUB_TOKEN not set. API rate limits may apply.")
        print()
    
    try:
        # Analyze container workflow
        print("📊 Analyzing Container Build Workflow...")
        print("-" * 70)
        
        runs = get_workflow_runs("container.yml", days=30)
        metrics = analyze_build_performance(runs)
        cache_metrics = calculate_cache_efficiency(runs)
        
        print(f"\n📈 Build Metrics (Last 30 days)")
        print(f"  Total Runs:        {metrics['total_runs']}")
        print(f"  Successful:        {metrics['successful_runs']}")
        print(f"  Failed:            {metrics['failed_runs']}")
        print(f"  Success Rate:      {metrics['success_rate']}%")
        print()
        print(f"⏱️  Duration Metrics")
        print(f"  Average:           {metrics['avg_duration_minutes']} minutes")
        print(f"  Minimum:           {metrics['min_duration_minutes']} minutes")
        print(f"  Maximum:           {metrics['max_duration_minutes']} minutes")
        print()
        print(f"💾 Cache Performance")
        print(f"  Estimated Hit Rate: {cache_metrics['cache_hit_rate_estimate']}%")
        print(f"  Fast Builds:        {cache_metrics['fast_builds']}/{cache_metrics['total_builds']}")
        print()
        
        # Recommendations
        recommendations = generate_recommendations(metrics)
        print("💡 Recommendations")
        for rec in recommendations:
            print(f"  {rec}")
        print()
        
        # Output JSON for automation
        output = {
            "metrics": metrics,
            "cache": cache_metrics,
            "recommendations": recommendations,
            "generated_at": datetime.now().isoformat()
        }
        
        with open("build-analytics.json", "w") as f:
            json.dump(output, f, indent=2)
        
        print("✅ Analysis complete. Results saved to build-analytics.json")
        
    except Exception as e:
        print(f"❌ Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
