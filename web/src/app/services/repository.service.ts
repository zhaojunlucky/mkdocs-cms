import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AuthService } from '../auth/auth.service';

export interface Repository {
  id: number;
  name: string;
  url: string;
  local_path: string;
  user_id: string;
  created_at: string;
  updated_at: string;
}

export interface Collection {
  id: number;
  name: string;
  path: string;
  repo_id: number;
  created_at: string;
  updated_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class RepositoryService {
  private apiUrl = 'http://localhost:8080/api';

  constructor(
    private http: HttpClient,
    private authService: AuthService
  ) { }

  // Get all repositories for the current user
  getUserRepositories(): Observable<Repository[]> {
    const headers = this.getAuthHeaders();
    const userId = this.authService.currentUser?.id;
    return this.http.get<Repository[]>(`${this.apiUrl}/v1/users/repos/${userId}`, { headers });
  }

  // Get a specific repository by ID
  getRepository(id: number): Observable<Repository> {
    const headers = this.getAuthHeaders();
    return this.http.get<Repository>(`${this.apiUrl}/repos/${id}`, { headers });
  }

  // Get collections for a repository
  getRepositoryCollections(repoId: number): Observable<Collection[]> {
    const headers = this.getAuthHeaders();
    return this.http.get<Collection[]>(`${this.apiUrl}/repos/${repoId}/collections`, { headers });
  }

  // Create a new repository
  createRepository(repository: Partial<Repository>): Observable<Repository> {
    const headers = this.getAuthHeaders();
    return this.http.post<Repository>(`${this.apiUrl}/repos`, repository, { headers });
  }

  // Update a repository
  updateRepository(id: number, repository: Partial<Repository>): Observable<Repository> {
    const headers = this.getAuthHeaders();
    return this.http.put<Repository>(`${this.apiUrl}/repos/${id}`, repository, { headers });
  }

  // Delete a repository
  deleteRepository(id: number): Observable<void> {
    const headers = this.getAuthHeaders();
    return this.http.delete<void>(`${this.apiUrl}/repos/${id}`, { headers });
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
