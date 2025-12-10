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
} from 'antd';
import {
  EyeOutlined,
  StopOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { Build, Pipeline } from '../types';
import { apiService } from '../services/api';

const { Text } = Typography;
const { Search } = Input;
const { Option } = Select;

const Builds: React.FC = () => {
  const [builds, setBuilds] = useState<Build[]>([]);
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [loading, setLoading] = useState(false);
  const [logsModalVisible, setLogsModalVisible] = useState(false);
  const [selectedBuild, setSelectedBuild] = useState<Build | null>(null);
  const [buildLogs, setBuildLogs] = useState('');
  const [logsLoading, setLogsLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedPipeline, setSelectedPipeline] = useState<number | undefined>();

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000); // 每10秒刷新一次
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [buildsRes, pipelinesRes] = await Promise.all([
        apiService.getBuilds(selectedPipeline),
        apiService.getPipelines(),
      ]);
      setBuilds(buildsRes.builds || []);
      setPipelines(pipelinesRes.pipelines || []);
    } catch (error) {
      console.error('Failed to fetch builds:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleViewLogs = async (build: Build) => {
    try {
      setSelectedBuild(build);
      setLogsLoading(true);
      setLogsModalVisible(true);
      const response = await apiService.getBuildLogs(build.id);
      setBuildLogs(response.logs || '暂无日志');
    } catch (error) {
      setBuildLogs('获取日志失败');
    } finally {
      setLogsLoading(false);
    }
  };

  const handleCancelBuild = async (build: Build) => {
    try {
      await apiService.cancelBuild(build.id);
      fetchData();
    } catch (error) {
      console.error('Failed to cancel build:', error);
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
        return '运行中';
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

  const filteredBuilds = builds.filter(build => {
    const matchesSearch = searchText === '' ||
      build.pipeline?.name.toLowerCase().includes(searchText.toLowerCase()) ||
      build.branch.toLowerCase().includes(searchText.toLowerCase()) ||
      build.tag?.toLowerCase().includes(searchText.toLowerCase());

    const matchesPipeline = !selectedPipeline || build.pipeline_id === selectedPipeline;

    return matchesSearch && matchesPipeline;
  });

  const columns = [
    {
      title: '流水线',
      dataIndex: ['pipeline', 'name'],
      key: 'pipeline',
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: '分支/标签',
      dataIndex: 'branch',
      key: 'branch',
      render: (branch: string, record: Build) => (
        <Space>
          <Tag color="blue">{record.tag || branch}</Tag>
          {record.commit_hash && (
            <Text type="secondary" code>
              {record.commit_hash.substring(0, 7)}
            </Text>
          )}
        </Space>
      ),
    },
    {
      title: '镜像',
      key: 'image',
      render: (_: any, record: Build) => (
        <Text code>
          {record.image_name}:{record.image_tag}
        </Text>
      ),
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
      title: '开始时间',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (time: string) => time ? new Date(time).toLocaleString() : '-',
    },
    {
      title: '耗时',
      key: 'duration',
      render: (_: any, record: Build) => {
        if (record.started_at && record.completed_at) {
          const start = new Date(record.started_at).getTime();
          const end = new Date(record.completed_at).getTime();
          const duration = Math.round((end - start) / 1000);
          return `${duration}s`;
        }
        return record.status === 'running' ? '运行中' : '-';
      },
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Build) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleViewLogs(record)}
          >
            日志
          </Button>
          {record.status === 'running' && (
            <Button
              type="text"
              danger
              icon={<StopOutlined />}
              onClick={() => handleCancelBuild(record)}
            >
              取消
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="构建记录"
      extra={
        <Space>
          <Select
            placeholder="选择流水线"
            allowClear
            style={{ width: 200 }}
            value={selectedPipeline}
            onChange={setSelectedPipeline}
          >
            {pipelines.map(pipeline => (
              <Option key={pipeline.id} value={pipeline.id}>
                {pipeline.name}
              </Option>
            ))}
          </Select>
          <Search
            placeholder="搜索流水线或分支"
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
        dataSource={filteredBuilds}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
      />

      <Modal
        title={`构建日志 - ${selectedBuild?.pipeline?.name}`}
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
            {buildLogs}
          </div>
        )}
      </Modal>
    </Card>
  );
};

export default Builds;