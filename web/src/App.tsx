import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from 'antd';
import { AuthProvider, useAuth } from './hooks/useAuth';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Projects from './pages/Projects';
import Pipelines from './pages/Pipelines';
import Builds from './pages/Builds';
import Deployments from './pages/Deployments';
import Settings from './pages/Settings';
import AppLayout from './components/Layout/AppLayout';

const { Content } = Layout;

function AppRoutes() {
  const { user, loading } = useAuth();

  if (loading) {
    return <div>加载中...</div>;
  }

  if (!user) {
    return <Login />;
  }

  return (
    <AppLayout>
      <Content>
        <Routes>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/projects" element={<Projects />} />
          <Route path="/projects/:id" element={<Projects />} />
          <Route path="/pipelines" element={<Pipelines />} />
          <Route path="/pipelines/:id" element={<Pipelines />} />
          <Route path="/builds" element={<Builds />} />
          <Route path="/builds/:id" element={<Builds />} />
          <Route path="/deployments" element={<Deployments />} />
          <Route path="/deployments/:id" element={<Deployments />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </Content>
    </AppLayout>
  );
}

function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  );
}

export default App;