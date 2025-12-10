import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Avatar,
  Upload,
  message,
  Tabs,
  Space,
  Divider,
  Typography,
} from 'antd';
import { UserOutlined, UploadOutlined, LockOutlined } from '@ant-design/icons';
import { useAuth } from '../hooks/useAuth';
import { apiService } from '../services/api';

const { Title, Text } = Typography;
const { TabPane } = Tabs;

const Settings: React.FC = () => {
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();

  useEffect(() => {
    if (user) {
      profileForm.setFieldsValue({
        username: user.username,
        email: user.email,
        avatar: user.avatar,
      });
    }
  }, [user, profileForm]);

  const handleProfileSubmit = async (values: any) => {
    try {
      setLoading(true);
      await apiService.updateProfile(values);
      message.success('个人资料更新成功');
      // 重新获取用户信息
      // TODO: 刷新用户状态
    } catch (error) {
      message.error('个人资料更新失败');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordSubmit = async (values: any) => {
    try {
      setPasswordLoading(true);
      // TODO: 实现修改密码 API
      message.success('密码修改成功');
      passwordForm.resetFields();
    } catch (error) {
      message.error('密码修改失败');
    } finally {
      setPasswordLoading(false);
    }
  };

  const uploadProps = {
    name: 'file',
    action: '/api/v1/users/avatar',
    headers: {
      authorization: `Bearer ${localStorage.getItem('token')}`,
    },
    beforeUpload: (file: File) => {
      const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
      if (!isJpgOrPng) {
        message.error('只能上传 JPG/PNG 文件!');
        return false;
      }
      const isLt2M = file.size / 1024 / 1024 < 2;
      if (!isLt2M) {
        message.error('图片必须小于 2MB!');
        return false;
      }
      return false; // 阻止自动上传
    },
    onChange(info: any) {
      if (info.file.status === 'done') {
        message.success('头像上传成功');
      }
    },
  };

  return (
    <div>
      <Title level={2}>系统设置</Title>

      <Tabs defaultActiveKey="profile">
        <TabPane tab="个人资料" key="profile">
          <Card title="个人信息">
            <div style={{ display: 'flex', marginBottom: 24 }}>
              <Avatar
                size={100}
                src={user?.avatar}
                icon={<UserOutlined />}
                style={{ marginRight: 24 }}
              />
              <div>
                <Title level={4}>{user?.username}</Title>
                <Text type="secondary">{user?.email}</Text>
                <div style={{ marginTop: 12 }}>
                  <Upload {...uploadProps} showUploadList={false}>
                    <Button icon={<UploadOutlined />}>更换头像</Button>
                  </Upload>
                </div>
              </div>
            </div>

            <Divider />

            <Form
              form={profileForm}
              layout="vertical"
              onFinish={handleProfileSubmit}
              style={{ maxWidth: 600 }}
            >
              <Form.Item
                name="username"
                label="用户名"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input disabled />
              </Form.Item>

              <Form.Item
                name="email"
                label="邮箱地址"
                rules={[
                  { required: true, message: '请输入邮箱地址' },
                  { type: 'email', message: '请输入有效的邮箱地址' }
                ]}
              >
                <Input />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  保存修改
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </TabPane>

        <TabPane tab="修改密码" key="password">
          <Card title="修改密码">
            <Form
              form={passwordForm}
              layout="vertical"
              onFinish={handlePasswordSubmit}
              style={{ maxWidth: 600 }}
            >
              <Form.Item
                name="currentPassword"
                label="当前密码"
                rules={[{ required: true, message: '请输入当前密码' }]}
              >
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>

              <Form.Item
                name="newPassword"
                label="新密码"
                rules={[
                  { required: true, message: '请输入新密码' },
                  { min: 6, message: '密码至少6个字符' }
                ]}
              >
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>

              <Form.Item
                name="confirmPassword"
                label="确认新密码"
                dependencies={['newPassword']}
                rules={[
                  { required: true, message: '请确认新密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('newPassword') === value) {
                        return Promise.resolve();
                      }
                      return Promise.reject(new Error('两次输入的密码不一致'));
                    },
                  }),
                ]}
              >
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>

              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit" loading={passwordLoading}>
                    修改密码
                  </Button>
                  <Button onClick={() => passwordForm.resetFields()}>
                    重置
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </TabPane>

        <TabPane tab="系统信息" key="system">
          <Card title="系统信息">
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text strong>系统版本：</Text>
                <Text>v1.0.0</Text>
              </div>
              <div>
                <Text strong>构建时间：</Text>
                <Text>2024-01-01 00:00:00</Text>
              </div>
              <div>
                <Text strong>Go 版本：</Text>
                <Text>1.24</Text>
              </div>
              <div>
                <Text strong>数据库版本：</Text>
                <Text>PostgreSQL 15</Text>
              </div>
              <div>
                <Text strong>Redis 版本：</Text>
                <Text>7.0</Text>
              </div>
            </Space>
          </Card>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default Settings;