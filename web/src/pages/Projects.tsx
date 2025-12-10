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
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  BranchesOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { Project } from '../types';
import { apiService } from '../services/api';

const { Option } = Select;
const { TextArea } = Input;

const Projects: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    try {
      setLoading(true);
      const response = await apiService.getProjects();
      setProjects(response.projects || []);
    } catch (error) {
      message.error('获取项目列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingProject(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (project: Project) => {
    setEditingProject(project);
    form.setFieldsValue(project);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await apiService.deleteProject(id);
      message.success('项目删除成功');
      fetchProjects();
    } catch (error) {
      message.error('项目删除失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      if (editingProject) {
        await apiService.updateProject(editingProject.id, values);
        message.success('项目更新成功');
      } else {
        await apiService.createProject(values);
        message.success('项目创建成功');
      }
      setModalVisible(false);
      fetchProjects();
    } catch (error) {
      message.error(editingProject ? '项目更新失败' : '项目创建失败');
    }
  };

  const columns = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Project) => (
        <Button
          type="link"
          onClick={() => navigate(`/projects/${record.id}`)}
        >
          {text}
        </Button>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Git 地址',
      dataIndex: 'git_url',
      key: 'git_url',
      ellipsis: true,
    },
    {
      title: 'Git 提供商',
      dataIndex: 'git_provider',
      key: 'git_provider',
      render: (provider: string) => (
        <Tag color={provider === 'github' ? 'black' : provider === 'gitlab' ? 'orange' : 'red'}>
          {provider.toUpperCase()}
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
      render: (_: any, record: Project) => (
        <Space>
          <Button
            type="text"
            icon={<BranchesOutlined />}
            onClick={() => navigate(`/pipelines?projectId=${record.id}`)}
          >
            流水线
          </Button>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个项目吗？"
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
      title="项目管理"
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新建项目
        </Button>
      }
    >
      <Table
        dataSource={projects}
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
        title={editingProject ? '编辑项目' : '新建项目'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="name"
            label="项目名称"
            rules={[{ required: true, message: '请输入项目名称' }]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="项目描述"
          >
            <TextArea placeholder="请输入项目描述" rows={3} />
          </Form.Item>

          <Form.Item
            name="git_url"
            label="Git 地址"
            rules={[
              { required: true, message: '请输入 Git 地址' },
              { type: 'url', message: '请输入有效的 Git 地址' }
            ]}
          >
            <Input placeholder="https://github.com/username/repo.git" />
          </Form.Item>

          <Form.Item
            name="git_provider"
            label="Git 提供商"
            rules={[{ required: true, message: '请选择 Git 提供商' }]}
          >
            <Select placeholder="请选择 Git 提供商">
              <Option value="github">GitHub</Option>
              <Option value="gitlab">GitLab</Option>
              <Option value="gitee">Gitee</Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingProject ? '更新' : '创建'}
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

export default Projects;