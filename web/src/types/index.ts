export interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  avatar?: string;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: number;
  name: string;
  description: string;
  git_url: string;
  git_provider: string;
  owner_id: number;
  created_at: string;
  updated_at: string;
  owner?: User;
  pipelines?: Pipeline[];
}

export interface Pipeline {
  id: number;
  name: string;
  description: string;
  project_id: number;
  config: string;
  status: string;
  created_at: string;
  updated_at: string;
  project?: Project;
  builds?: Build[];
  triggers?: PipelineTrigger[];
}

export interface PipelineTrigger {
  id: number;
  pipeline_id: number;
  type: string;
  branch?: string;
  tag?: string;
  schedule?: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Build {
  id: number;
  pipeline_id: number;
  commit_hash: string;
  branch: string;
  tag?: string;
  status: 'pending' | 'running' | 'success' | 'failed' | 'cancelled';
  logs: string;
  image_name: string;
  image_tag: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  pipeline?: Pipeline;
  deployments?: Deployment[];
}

export interface Deployment {
  id: number;
  build_id: number;
  environment: 'dev' | 'staging' | 'prod';
  status: 'pending' | 'running' | 'success' | 'failed' | 'cancelled';
  replicas: number;
  namespace: string;
  service_name: string;
  ingress_host?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  build?: Build;
}

export interface EnvironmentVariable {
  id: number;
  project_id: number;
  key: string;
  value: string;
  secret: boolean;
  scope: 'build' | 'deploy' | 'all';
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  message: string;
  user: User;
  token: string;
}

export interface ApiResponse<T = any> {
  message: string;
  data?: T;
}