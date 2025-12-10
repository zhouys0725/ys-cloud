import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { message } from 'antd';

const API_BASE_URL = process.env.REACT_APP_API_URL || '/api/v1';

class ApiService {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Request interceptor
    this.api.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor
    this.api.interceptors.response.use(
      (response: AxiosResponse) => {
        return response;
      },
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          window.location.href = '/login';
          return Promise.reject(error);
        }

        const errorMessage = error.response?.data?.error || '请求失败';
        message.error(errorMessage);
        return Promise.reject(error);
      }
    );
  }

  // Auth methods
  async login(username: string, password: string) {
    const response = await this.api.post('/auth/login', { username, password });
    return response.data;
  }

  async register(username: string, email: string, password: string) {
    const response = await this.api.post('/auth/register', { username, email, password });
    return response.data;
  }

  async getProfile() {
    const response = await this.api.get('/users/profile');
    return response.data;
  }

  async updateProfile(data: any) {
    const response = await this.api.put('/users/profile', data);
    return response.data;
  }

  // Project methods
  async getProjects() {
    const response = await this.api.get('/projects');
    return response.data;
  }

  async getProject(id: number) {
    const response = await this.api.get(`/projects/${id}`);
    return response.data;
  }

  async createProject(data: any) {
    const response = await this.api.post('/projects', data);
    return response.data;
  }

  async updateProject(id: number, data: any) {
    const response = await this.api.put(`/projects/${id}`, data);
    return response.data;
  }

  async deleteProject(id: number) {
    const response = await this.api.delete(`/projects/${id}`);
    return response.data;
  }

  // Pipeline methods
  async getPipelines(projectId?: number) {
    const params = projectId ? { projectId } : {};
    const response = await this.api.get('/pipelines', { params });
    return response.data;
  }

  async getPipeline(id: number) {
    const response = await this.api.get(`/pipelines/${id}`);
    return response.data;
  }

  async createPipeline(data: any) {
    const response = await this.api.post('/pipelines', data);
    return response.data;
  }

  async updatePipeline(id: number, data: any) {
    const response = await this.api.put(`/pipelines/${id}`, data);
    return response.data;
  }

  async deletePipeline(id: number) {
    const response = await this.api.delete(`/pipelines/${id}`);
    return response.data;
  }

  async runPipeline(id: number) {
    const response = await this.api.post(`/pipelines/${id}/run`);
    return response.data;
  }

  // Build methods
  async getBuilds(pipelineId?: number) {
    const params = pipelineId ? { pipelineId } : {};
    const response = await this.api.get('/builds', { params });
    return response.data;
  }

  async getBuild(id: number) {
    const response = await this.api.get(`/builds/${id}`);
    return response.data;
  }

  async getBuildLogs(id: number) {
    const response = await this.api.get(`/builds/${id}/logs`);
    return response.data;
  }

  async cancelBuild(id: number) {
    const response = await this.api.post(`/builds/${id}/cancel`);
    return response.data;
  }

  // Deployment methods
  async getDeployments(buildId?: number, environment?: string) {
    const params: any = {};
    if (buildId) params.buildId = buildId;
    if (environment) params.environment = environment;

    const response = await this.api.get('/deployments', { params });
    return response.data;
  }

  async getDeployment(id: number) {
    const response = await this.api.get(`/deployments/${id}`);
    return response.data;
  }

  async getDeploymentLogs(id: number) {
    const response = await this.api.get(`/deployments/${id}/logs`);
    return response.data;
  }

  async rollbackDeployment(id: number) {
    const response = await this.api.post(`/deployments/${id}/rollback`);
    return response.data;
  }
}

export const apiService = new ApiService();