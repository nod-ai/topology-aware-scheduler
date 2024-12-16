import streamlit as st
import plotly.graph_objs as go
import plotly.express as px
import numpy as np
from dataclasses import dataclass
from typing import List, Dict
from enum import Enum
import random
import pandas as pd
from datetime import datetime
import time

class NodeStatus(Enum):
    FREE = 0
    OCCUPIED = 1

@dataclass
class Job:
    id: str
    nodes_required: int
    duration: int
    current_runtime: int = 0

class ClusterState:
    def __init__(self, total_nodes: int = 128, gpus_per_node: int = 4, nodes_per_leaf: int = 4):
        self.total_nodes = total_nodes
        self.gpus_per_node = gpus_per_node
        self.nodes_per_leaf = nodes_per_leaf
        self.leaf_switches = total_nodes // nodes_per_leaf
        self.node_status = np.zeros(total_nodes, dtype=int)
        self.running_jobs: Dict[str, Job] = {}

class TopologyAwareSchedulerWithMetrics:
    def __init__(self, cluster: ClusterState):
        self.cluster = cluster
        self.job_queue: List[Job] = []
        self.metrics = pd.DataFrame(columns=[
            'timestamp', 'gpu_utilization', 'queue_length', 
            'active_jobs', 'successful_placements', 'failed_placements'
        ])
        self.successful_placements = 0
        self.failed_placements = 0
    
    def submit_job(self, job: Job) -> bool:
        if self._can_schedule_job(job):
            self._allocate_resources(job)
            self.cluster.running_jobs[job.id] = job
            self.successful_placements += 1
            return True
        self.job_queue.append(job)
        self.failed_placements += 1
        return False
    
    def _can_schedule_job(self, job: Job) -> bool:
        if job.nodes_required > self.cluster.total_nodes:
            return False
        
        leaves_needed = (job.nodes_required + self.cluster.nodes_per_leaf - 1) // self.cluster.nodes_per_leaf
        
        for i in range(self.cluster.leaf_switches - leaves_needed + 1):
            start_node = i * self.cluster.nodes_per_leaf
            end_node = start_node + (leaves_needed * self.cluster.nodes_per_leaf)
            
            if np.sum(self.cluster.node_status[start_node:end_node]) == 0:
                return True
        return False
    
    def _allocate_resources(self, job: Job) -> None:
        leaves_needed = (job.nodes_required + self.cluster.nodes_per_leaf - 1) // self.cluster.nodes_per_leaf
        
        for i in range(self.cluster.leaf_switches - leaves_needed + 1):
            start_node = i * self.cluster.nodes_per_leaf
            end_node = start_node + (leaves_needed * self.cluster.nodes_per_leaf)
            
            if np.sum(self.cluster.node_status[start_node:end_node]) == 0:
                self.cluster.node_status[start_node:start_node + job.nodes_required] = NodeStatus.OCCUPIED.value
                break
    
    def update_metrics(self):
        now = datetime.now()
        gpu_util = np.sum(self.cluster.node_status) / len(self.cluster.node_status)
        
        new_metrics = pd.DataFrame({
            'timestamp': [now],
            'gpu_utilization': [gpu_util * 100],
            'queue_length': [len(self.job_queue)],
            'active_jobs': [len(self.cluster.running_jobs)],
            'successful_placements': [self.successful_placements],
            'failed_placements': [self.failed_placements]
        })
        self.metrics = pd.concat([self.metrics, new_metrics], ignore_index=True)
    
    def process_completed_jobs(self):
        completed_jobs = []
        for job_id, job in self.cluster.running_jobs.items():
            job.current_runtime += 1
            if job.current_runtime >= job.duration:
                completed_jobs.append(job_id)
        
        for job_id in completed_jobs:
            job = self.cluster.running_jobs[job_id]
            start_node = np.where(self.cluster.node_status == NodeStatus.OCCUPIED.value)[0][0]
            self.cluster.node_status[start_node:start_node + job.nodes_required] = NodeStatus.FREE.value
            del self.cluster.running_jobs[job_id]

def initialize_session_state():
    if 'cluster' not in st.session_state:
        st.session_state.cluster = ClusterState()
        st.session_state.scheduler = TopologyAwareSchedulerWithMetrics(st.session_state.cluster)
        st.session_state.start_time = datetime.now()
        st.session_state.iteration = 0

def render_dashboard():
    st.title('GPU Cluster Scheduler Dashboard')
    st.markdown("""
    This dashboard simulates a topology-aware GPU cluster scheduler.
    - Each cell in the topology view represents a node with 4 GPUs
    - Darker cells indicate nodes in use
    - Columns represent leaf switch domains
    """)
    
    initialize_session_state()
    st.session_state.iteration += 1
    
    with st.sidebar:
        st.header('Simulation Controls')
        job_probability = st.slider('Job Arrival Rate', 0.0, 1.0, 0.3)
        job_sizes = st.multiselect('Job Sizes (nodes)', [2, 4, 8, 16], default=[2, 4, 8, 16])
        update_interval = st.slider('Update Interval (seconds)', 0.1, 2.0, 1.0)

    if random.random() < job_probability:
        job = Job(
            id=f"job_{st.session_state.iteration}",
            nodes_required=random.choice(job_sizes),
            duration=random.randint(5, 15)
        )
        st.session_state.scheduler.submit_job(job)
    
    st.session_state.scheduler.process_completed_jobs()
    st.session_state.scheduler.update_metrics()
    
    col1, col2, col3, col4 = st.columns(4)
    col1.metric("Active Jobs", len(st.session_state.scheduler.cluster.running_jobs))
    col2.metric("Queue Length", len(st.session_state.scheduler.job_queue))
    col3.metric("GPU Utilization", 
                f"{st.session_state.scheduler.metrics['gpu_utilization'].iloc[-1]:.1f}%")
    success_rate = ((st.session_state.scheduler.successful_placements / 
                    (st.session_state.scheduler.successful_placements + 
                     st.session_state.scheduler.failed_placements) * 100)
                    if (st.session_state.scheduler.successful_placements + 
                        st.session_state.scheduler.failed_placements) > 0 else 0)
    col4.metric("Success Rate", f"{success_rate:.1f}%")
    
    node_matrix = st.session_state.scheduler.cluster.node_status.reshape(-1, 
        st.session_state.scheduler.cluster.nodes_per_leaf)
    
    hover_text = [[f"Leaf Switch: {j+1}<br>Node: {i+1}<br>Status: {'In Use' if node_matrix[i][j] else 'Free'}"
                  for j in range(node_matrix.shape[1])]
                 for i in range(node_matrix.shape[0])]
    
    fig_topology = go.Figure(data=go.Heatmap(
        z=node_matrix,
        text=hover_text,
        hoverongaps=False,
        colorscale=[[0, 'lightgrey'], [1, 'darkblue']],
        showscale=False
    ))
    
    fig_topology.update_layout(
        title='Cluster Topology State',
        xaxis_title="Leaf Switch Domains",
        yaxis_title="Nodes",
        height=400
    )
    
    st.plotly_chart(fig_topology, use_container_width=True)
    
    col1, col2 = st.columns(2)
    
    with col1:
        fig_util = px.line(
            st.session_state.scheduler.metrics.tail(50), 
            x='timestamp', 
            y='gpu_utilization',
            title='GPU Utilization Over Time'
        )
        fig_util.update_layout(
            xaxis_title="Time",
            yaxis_title="Utilization (%)",
            height=300
        )
        st.plotly_chart(fig_util, use_container_width=True)
    
    with col2:
        fig_queue = px.line(
            st.session_state.scheduler.metrics.tail(50), 
            x='timestamp', 
            y=['queue_length', 'active_jobs'],
            title='Workload Distribution'
        )
        fig_queue.update_layout(
            xaxis_title="Time",
            yaxis_title="Number of Jobs",
            height=300,
            legend_title="Job Status"
        )
        fig_queue.data[0].name = "Waiting Jobs"
        fig_queue.data[1].name = "Running Jobs"
        st.plotly_chart(fig_queue, use_container_width=True)
    
    time.sleep(update_interval)
    st.rerun()

if __name__ == '__main__':
    render_dashboard()
