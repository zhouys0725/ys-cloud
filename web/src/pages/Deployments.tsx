import React, { useState, useEffect } from 'react';
import {
  Table,
  Card,
  Tag,
  Button,
  Space,
  Modal,
  Typography,
  Spin,
  Input,
  Select,
  Progress,
  Statistic,
  Row,
  Col,
} from 'antd';
import {
  EyeOutlined,
  RollbackOutlined,
  ReloadOutlined,
  EnvironmentOutlined,
  CloudServerOutlined,
} from '@ant-design/icons';
import { Deployment, Build } from '../types';
import { apiService } from '../services/api';

const { Text, Title } = Typography;
const { Search } = Input;
const { Option } = Select;

const Deployments: React.FC = () => {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(false);
  const [logsModalVisible, setLogsModalVisible] = useState(false);
  const [selectedDeployment, setSelectedDeployment] = useState<Deployment | null>(null);
  const [deploymentLogs, setDeploymentLogs] = useState('');
  const [logsLoading, setLogsLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedEnvironment, setSelectedEnvironment] = useState<string | undefined>();

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 15000); // 每15秒刷新一次
    return () => clearInterval(interval);
  }, [selectedEnvironment]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await apiService.getDeployments(undefined, selectedEnvironment);
      setDeployments(response.deployments || []);
    } catch (error) {
      console.error('Failed to fetch deployments:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleViewLogs = async (deployment: Deployment) => {
    try {
      setSelectedDeployment(deployment);
      setLogsLoading(true);
      setLogsModalVisible(true);
      const response = await apiService.getDeploymentLogs(deployment.id);
      setDeploymentLogs(response.logs || '暂无日志');
    } catch (error) {
      setDeploymentLogs('获取日志失败');
    } finally {
      setLogsLoading(false);
    }
  };

  const handleRollback = async (deployment: Deployment) => {
    try {
      await apiService.rollbackDeployment(deployment.id);
      fetchData();
    } catch (error) {
      console.error('Failed to rollback deployment:', error);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'green';
      case 'running':
        return 'blue';
      case 'failed':
        return 'red';
      case 'cancelled':
        return 'default';
      case 'pending':
        return 'orange';
      default:
        return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'success':
        return '成功';
      case 'running':
        return '部署中';
      case 'failed':
        return '失败';
      case 'cancelled':
        return '已取消';
      case 'pending':
        return '等待中';
      default:
        return status;
    }
  };

  const getEnvironmentColor = (env: string) => {
    switch (env) {
      case 'prod':
        return 'red';
      case 'staging':
        return 'orange';
      case 'dev':
        return 'blue';
      default:
        return 'default';
    }
  };

  const filteredDeployments = deployments.filter(deployment => {
    const matchesSearch = searchText === '' ||
      deployment.service_name.toLowerCase().includes(searchText.toLowerCase()) ||
      deployment.namespace.toLowerCase().includes(searchText.toLowerCase()) ||
      deployment.ingress_host?.toLowerCase().includes(searchText.toLowerCase());

    const matchesEnvironment = !selectedEnvironment || deployment.environment === selectedEnvironment;

    return matchesSearch && matchesEnvironment;
  });

  // 统计数据
  const stats = {
    total: deployments.length,
    success: deployments.filter(d => d.status === 'success').length,
    running: deployments.filter(d => d.status === 'running').length,
    failed: deployments.filter(d => d.status === 'failed').length,
  };

  const columns = [
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      render: (env: string) => (
        <Tag color={getEnvironmentColor(env)}>
          {env.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '服务名',
      dataIndex: 'service_name',
      key: 'service_name',
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: '命名空间',
      dataIndex: 'namespace',
      key: 'namespace',
      render: (text: string) => <Text code>{text}</Text>,
    },
    {
      title: '副本数',
      dataIndex: 'replicas',
      key: 'replicas',
      render: (replicas: number) => (
        <Tag color="blue">{replicas}</Tag>
      ),
    },
    {
      title: '构建ID',
      dataIndex: ['build', 'id'],
      key: 'build_id',
      render: (id: number) => <Text code>#{id}</Text>,
    },
    {
      title: '域名',
      dataIndex: 'ingress_host',
      key: 'ingress_host',
      render: (host: string) => host ? (
        <a href={`http://${host}`} target="_blank" rel="noopener noreferrer">
          {host}
        </a>
      ) : '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: '部署时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Deployment) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleViewLogs(record)}
          >
            日志
          </Button>
          {record.status === 'success' && (
            <Button
              type="text"
              icon={<RollbackOutlined />}
              onClick={() => handleRollback(record)}
            >
              回滚
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总部署数"
              value={stats.total}
              prefix={<CloudServerOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功部署"
              value={stats.success}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="部署中"
              value={stats.running}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="失败部署"
              value={stats.failed}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title="部署管理"
        extra={
          <Space>
            <Select
              placeholder="选择环境"
              allowClear
              style={{ width: 120 }}
              value={selectedEnvironment}
              onChange={setSelectedEnvironment}
            >
              <Option value="dev">开发环境</Option>
              <Option value="staging">测试环境</Option>
              <Option value="prod">生产环境</Option>
            </Select>
            <Search
              placeholder="搜索服务名或域名"
              allowClear
              style={{ width: 250 }}
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
            />
            <Button icon={<ReloadOutlined />} onClick={fetchData}>
              刷新
            </Button>
          </Space>
        }
      >
        <Table
          dataSource={filteredDeployments}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      <Modal
        title={`部署日志 - ${selectedDeployment?.service_name}`}
        open={logsModalVisible}
        onCancel={() => setLogsModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setLogsModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={800}
      >
        {logsLoading ? (
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <Spin size="large" />
            <div style={{ marginTop: 16 }}>加载日志中...</div>
          </div>
        ) : (
          <div
            style={{
              backgroundColor: '#001529',
              color: '#fff',
              padding: '16px',
              borderRadius: '4px',
              fontFamily: 'monospace',
              fontSize: '12px',
              height: '400px',
              overflow: 'auto',
              whiteSpace: 'pre-wrap',
            }}
          >
            {deploymentLogs}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Deployments;