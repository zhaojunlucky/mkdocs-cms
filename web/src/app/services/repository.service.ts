import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
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
    const userId = this.authService.currentUser?.id;
    return this.http.get<ArrayResponse<Repository>>(`${this.apiUrl}/v1/users/repos/${userId}`);
  }

  // Get a specific repository by ID
  getRepository(id: number | string): Observable<Repository> {
    return this.http.get<Repository>(`${this.apiUrl}/v1/repos/${id}`);
  }

  // Get branches for a repository
  getRepositoryBranches(id: number): Observable<ArrayResponse<string>> {
    return this.http.get<ArrayResponse<string>>(`${this.apiUrl}/v1/repos/${id}/branches`);
  }

  // Get collections for a repository
  getRepositoryCollections(repoId: number|string): Observable<ArrayResponse<Collection>> {
    return this.http.get<ArrayResponse<Collection>>(`${this.apiUrl}/v1/collections/repo/${repoId}`);
  }

  // Update a repository
  updateRepository(id: number, repository: Partial<Repository>): Observable<Repository> {
    return this.http.put<Repository>(`${this.apiUrl}/v1/repos/${id}`, repository);
  }

  // Delete a repository
  deleteRepository(id: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/v1/repos/${id}`);
  }

  // Sync a repository
  syncRepository(id: number): Observable<SyncResponse> {
    return this.http.post<SyncResponse>(`${this.apiUrl}/v1/repos/${id}/sync`, {});
  }

  // Get a task by ID
  getTask(taskId: string): Observable<AsyncTask> {
    return this.http.get<AsyncTask>(`${this.apiUrl}/v1/tasks/${taskId}`);
  }
}
