import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AuthService } from '../auth/auth.service';
import {environment} from '../../environments/environment';
import {ArrayResponse} from '../shared/core/response';

export interface Repository {
  id: number;
  name: string;
  description?: string;
  remote_url?: string;
  branch?: string;
  local_path: string;
  user_id: string;
  status: string;
  last_sync_at?: string;
  error_msg?: string;
  created_at: string;
  updated_at: string;
  showMenu?: boolean;
  syncing?: boolean;
}

export interface CollectionFieldDefinition {
  type: string;
  name: string;
  label: string;
  required?: boolean;
  format?: string;
  list?: boolean;
  default?: any;
}

export interface Collection {
  id: number;
  name: string;
  label: string;
  description?: string;
  path: string;
  repo_id: number;
  created_at: string;
  updated_at: string;
  fields?: CollectionFieldDefinition[];
}

export interface AsyncTask {
  id: string;
  type: string;
  status: string;
  resource_id: string;
  user_id: string;
  message: string;
  progress: number;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  status: string;
  progress?: number;
  message?: string;
}

export interface SyncResponse {
  message: string;
  task_id: string;
}

@Injectable({
  providedIn: 'root'
})
export class RepositoryService {
  private apiUrl = environment.apiServer

  constructor(
    private http: HttpClient,
    private authService: AuthService
  ) { }

  // Get all repositories for the current user
  getUserRepositories(): Observable<ArrayResponse<Repository>> {
    const headers = this.getAuthHeaders();
    const userId = this.authService.currentUser?.id;
    return this.http.get<ArrayResponse<Repository>>(`${this.apiUrl}/v1/users/repos/${userId}`, { headers });
  }

  // Get a specific repository by ID
  getRepository(id: number | string): Observable<Repository> {
    const headers = this.getAuthHeaders();
    return this.http.get<Repository>(`${this.apiUrl}/v1/repos/${id}`, { headers });
  }

  // Get branches for a repository
  getRepositoryBranches(id: number): Observable<ArrayResponse<string>> {
    const headers = this.getAuthHeaders();
    return this.http.get<ArrayResponse<string>>(`${this.apiUrl}/v1/repos/${id}/branches`, { headers });
  }

  // Get collections for a repository
  getRepositoryCollections(repoId: number|string): Observable<ArrayResponse<Collection>> {
    const headers = this.getAuthHeaders();
    return this.http.get<ArrayResponse<Collection>>(`${this.apiUrl}/v1/collections/repo/${repoId}`, { headers });
  }

  // Update a repository
  updateRepository(id: number, repository: Partial<Repository>): Observable<Repository> {
    const headers = this.getAuthHeaders();
    return this.http.put<Repository>(`${this.apiUrl}/v1/repos/${id}`, repository, { headers });
  }

  // Delete a repository
  deleteRepository(id: number): Observable<void> {
    const headers = this.getAuthHeaders();
    return this.http.delete<void>(`${this.apiUrl}/v1/repos/${id}`, { headers });
  }

  // Sync a repository
  syncRepository(id: number): Observable<SyncResponse> {
    const headers = this.getAuthHeaders();
    return this.http.post<SyncResponse>(`${this.apiUrl}/v1/repos/${id}/sync`, {}, { headers });
  }

  // Get a task by ID
  getTask(taskId: string): Observable<AsyncTask> {
    const headers = this.getAuthHeaders();
    return this.http.get<AsyncTask>(`${this.apiUrl}/v1/tasks/${taskId}`, { headers });
  }

  // Helper method to get authentication headers
  private getAuthHeaders(): HttpHeaders {
    const token = this.authService.currentToken;
    return new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
  }
}
