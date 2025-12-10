import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Card,
  Space,
  Tag,
  message,
  Popconfirm,
  Tabs,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { useSearchParams } from 'react-router-dom';
import { Pipeline, Project } from '../types';
import { apiService } from '../services/api';

const { Option } = Select;
const { TextArea } = Input;
const { TabPane } = Tabs;

const Pipelines: React.FC = () => {
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingPipeline, setEditingPipeline] = useState<Pipeline | null>(null);
  const [activeTab, setActiveTab] = useState('list');
  const [form] = Form.useForm();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    fetchData();
  }, [searchParams]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const projectId = searchParams.get('projectId');
      const [pipelinesRes, projectsRes] = await Promise.all([
        apiService.getPipelines(projectId ? parseInt(projectId) : undefined),
        apiService.getProjects(),
      ]);
      setPipelines(pipelinesRes.pipelines || []);
      setProjects(projectsRes.projects || []);
    } catch (error) {
      message.error('获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingPipeline(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (pipeline: Pipeline) => {
    setEditingPipeline(pipeline);
    form.setFieldsValue(pipeline);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await apiService.deletePipeline(id);
      message.success('流水线删除成功');
      fetchData();
    } catch (error) {
      message.error('流水线删除失败');
    }
  };

  const handleRun = async (id: number) => {
    try {
      await apiService.runPipeline(id);
      message.success('流水线启动成功');
      fetchData();
    } catch (error) {
      message.error('流水线启动失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      if (editingPipeline) {
        await apiService.updatePipeline(editingPipeline.id, values);
        message.success('流水线更新成功');
      } else {
        await apiService.createPipeline(values);
        message.success('流水线创建成功');
      }
      setModalVisible(false);
      fetchData();
    } catch (error) {
      message.error(editingPipeline ? '流水线更新失败' : '流水线创建失败');
    }
  };

  const defaultPipelineConfig = `# 流水线配置示例
version: '1.0'

stages:
  - name: build
    image: golang:1.24
    commands:
      - go mod download
      - go build -o app .
    artifacts:
      - path: ./app

  - name: docker
    image: docker:latest
    commands:
      - docker build -t your-app:\${BUILD_NUMBER} .
      - docker push your-app:\${BUILD_NUMBER}

  - name: deploy
    image: kubectl:latest
    commands:
      - kubectl apply -f k8s/`;

  const columns = [
    {
      title: '流水线名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '项目',
      dataIndex: ['project', 'name'],
      key: 'project',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'default'}>
          {status === 'active' ? '激活' : '未激活'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Pipeline) => (
        <Space>
          <Button
            type="text"
            icon={<PlayCircleOutlined />}
            onClick={() => handleRun(record.id)}
          >
            运行
          </Button>
          <Button
            type="text"
            icon={<SettingOutlined />}
          >
            配置
          </Button>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这条流水线吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="text" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="流水线管理"
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新建流水线
        </Button>
      }
    >
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="流水线列表" key="list">
          <Table
            dataSource={pipelines}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={{
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 条记录`,
            }}
          />
        </TabPane>
      </Tabs>

      <Modal
        title={editingPipeline ? '编辑流水线' : '新建流水线'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ config: defaultPipelineConfig }}
        >
          <Form.Item
            name="name"
            label="流水线名称"
            rules={[{ required: true, message: '请输入流水线名称' }]}
          >
            <Input placeholder="请输入流水线名称" />
          </Form.Item>

          <Form.Item
            name="projectId"
            label="所属项目"
            rules={[{ required: true, message: '请选择项目' }]}
          >
            <Select placeholder="请选择项目">
              {projects.map(project => (
                <Option key={project.id} value={project.id}>
                  {project.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea placeholder="请输入流水线描述" rows={2} />
          </Form.Item>

          <Form.Item
            name="config"
            label="流水线配置"
            rules={[{ required: true, message: '请输入流水线配置' }]}
          >
            <TextArea
              placeholder="请输入流水线配置 (YAML 格式)"
              rows={12}
              style={{ fontFamily: 'monospace' }}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingPipeline ? '更新' : '创建'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default Pipelines;