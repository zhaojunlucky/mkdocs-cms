import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AuthService } from '../auth/auth.service';

export interface FileInfo {
  name: string;
  path: string;
  is_dir: boolean;
  size: number;
  mod_time: string;
  extension?: string;
  content?: string;
}

@Injectable({
  providedIn: 'root'
})
export class CollectionService {
  private apiUrl = 'http://localhost:8080/api';

  constructor(
    private http: HttpClient,
    private authService: AuthService
  ) { }

  // Get all files in a collection
  getCollectionFiles(repoId: number, collectionName: string): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/v1/repos-collections/${repoId}/collections/${collectionName}/files`, this.getAuthHeaders());
  }

  // Get files in a specific path within a collection
  getCollectionFilesInPath(repoId: number, collectionName: string, path: string): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/v1/repos-collections/${repoId}/collections/${collectionName}/browse?path=${path}`, this.getAuthHeaders());
  }

  // Get file content
  getFileContent(repoId: number, collectionName: string, path: string): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/v1/repos-collections/${repoId}/collections/${collectionName}/file?path=${path}`, this.getAuthHeaders());
  }

  // Update file content
  updateFileContent(repoId: number, collectionName: string, path: string, content: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/v1/repos-collections/${repoId}/collections/${collectionName}/file`, {
      path: path,
      content: content
    }, this.getAuthHeaders());
  }

  // Delete file
  deleteFile(repoId: number, collectionName: string, path: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/v1/repos-collections/${repoId}/collections/${collectionName}/file?path=${path}`, this.getAuthHeaders());
  }

  // Upload file
  uploadFile(repoId: number, collectionName: string, path: string, content: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/v1/repos-collections/${repoId}/collections/${collectionName}/file`, {
      path: path,
      content: content
    }, this.getAuthHeaders());
  }

  // Get file content as JSON
  getFileContentJSON(repoId: number, collectionName: string, path: string): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/v1/repos-collections/${repoId}/collections/${collectionName}/file/json?path=${path}`, this.getAuthHeaders());
  }

  // Helper method to get authentication headers
  private getAuthHeaders() {
    const token = this.authService.currentToken;
    return {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      })
    };
  }
}
