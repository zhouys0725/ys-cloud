import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Table, Tag, Progress, Spin } from 'antd';
import {
  ProjectOutlined,
  BranchesOutlined,
  BuildOutlined,
  DeploymentUnitOutlined,
} from '@ant-design/icons';
import { Project, Pipeline, Build, Deployment } from '../types';
import { apiService } from '../services/api';

const Dashboard: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [projects, setProjects] = useState<Project[]>([]);
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [builds, setBuilds] = useState<Build[]>([]);
  const [deployments, setDeployments] = useState<Deployment[]>([]);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [projectsRes, pipelinesRes, buildsRes, deploymentsRes] = await Promise.all([
        apiService.getProjects(),
        apiService.getPipelines(),
        apiService.getBuilds(),
        apiService.getDeployments(),
      ]);

      setProjects(projectsRes.projects || []);
      setPipelines(pipelinesRes.pipelines || []);
      setBuilds(buildsRes.builds || []);
      setDeployments(deploymentsRes.deployments || []);
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
      case 'completed':
        return 'green';
      case 'running':
        return 'blue';
      case 'failed':
        return 'red';
      case 'pending':
        return 'orange';
      default:
        return 'default';
    }
  };

  const buildColumns = [
    {
      title: '流水线',
      dataIndex: ['pipeline', 'name'],
      key: 'pipeline',
    },
    {
      title: '分支/标签',
      dataIndex: 'branch',
      key: 'branch',
      render: (branch: string, record: Build) => record.tag || branch,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '开始时间',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (time: string) => time ? new Date(time).toLocaleString() : '-',
    },
  ];

  const deploymentColumns = [
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      render: (env: string) => (
        <Tag color={env === 'prod' ? 'red' : env === 'staging' ? 'orange' : 'blue'}>
          {env.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '构建',
      dataIndex: ['build', 'id'],
      key: 'build',
    },
    {
      title: '副本数',
      dataIndex: 'replicas',
      key: 'replicas',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '部署时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => new Date(time).toLocaleString(),
    },
  ];

  const runningBuilds = builds.filter(b => b.status === 'running').length;
  const successfulDeployments = deployments.filter(d => d.status === 'success').length;

  if (loading) {
    return <Spin size="large" style={{ display: 'block', textAlign: 'center', marginTop: '100px' }} />;
  }

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总项目数"
              value={projects.length}
              prefix={<ProjectOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="流水线数"
              value={pipelines.length}
              prefix={<BranchesOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="运行中构建"
              value={runningBuilds}
              prefix={<BuildOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功部署"
              value={successfulDeployments}
              prefix={<DeploymentUnitOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Card title="最近构建" style={{ marginBottom: 16 }}>
            <Table
              dataSource={builds.slice(0, 10)}
              columns={buildColumns}
              pagination={false}
              rowKey="id"
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="最近部署">
            <Table
              dataSource={deployments.slice(0, 10)}
              columns={deploymentColumns}
              pagination={false}
              rowKey="id"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;