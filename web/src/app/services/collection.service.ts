import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AuthService } from '../auth/auth.service';
import { environment } from '../../environments/environment';
import {ArrayResponse} from '../shared/core/response';

export interface FileInfo {
  name: string;
  path: string;
  is_dir: boolean;
  size: number;
  mod_time: string;
  extension?: string;
  content?: string;
}

export interface CollectionField {
  type: 'string' | 'datetime' | 'boolean' | 'number';
  name: string;
  label: string;
  required?: boolean;
  format?: string;
  list?: boolean;
  default?: any;
}

export interface CollectionConfig {
  name: string;
  label: string;
  path: string;
  format: string;
  fields?: CollectionField[];
}

@Injectable({
  providedIn: 'root'
})
export class CollectionService {
  private apiUrl = environment.apiServer;

  constructor(
    private http: HttpClient,
    private authService: AuthService
  ) { }

  // Get all files in a collection
  getCollectionFiles(repoId: string | number, collectionName: string, path: string = ''): Observable<ArrayResponse<FileInfo>> {
    if (path) {
      return this.getCollectionFilesInPath(repoId, collectionName, path);
    }
    return this.http.get<ArrayResponse<FileInfo>>(`${this.apiUrl}/v1/collections/repo/${repoId}/${collectionName}/files`, this.getAuthHeaders());
  }

  // Get files in a specific path within a collection
  getCollectionFilesInPath(repoId: string | number, collectionName: string, path: string): Observable<ArrayResponse<FileInfo>> {
    return this.http.get<ArrayResponse<FileInfo>>(`${this.apiUrl}/v1/collections/repo/${repoId}/${collectionName}/files?path=${path}`, this.getAuthHeaders());
  }

  // Get file content
  getFileContent(repoId: string | number, collectionName: string, path: string): Observable<string> {
    return this.http.get(`${this.apiUrl}/v1/collections/repo/${repoId}/${collectionName}/files/content?path=${path}`, {
      ...this.getAuthHeaders(),
      responseType: 'text'
    });
  }

  // Update file content
  updateFileContent(repoId: string | number, collectionName: string, path: string, content: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/v1/collections/repo/${repoId}/${collectionName}/files/content`, {
      path: path,
      content: content
    }, this.getAuthHeaders());
  }

  // Delete file
  deleteFile(repoId: string | number, collectionName: string, path: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/v1/collections/repo/${repoId}/${collectionName}/files?path=${path}`, this.getAuthHeaders());
  }

  // Rename file
  renameFile(repoId: string | number, collectionName: string, oldPath: string, newPath: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/v1/collections/repo/${repoId}/${collectionName}/files/rename`, {
      oldPath: oldPath,
      newPath: newPath
    }, this.getAuthHeaders());
  }

  // Upload file
  uploadFile(repoId: string | number, collectionName: string, path: string, content: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/v1/collections/repo/${repoId}/${collectionName}/files/upload`, {
      path: path,
      content: content
    }, this.getAuthHeaders());
  }

  createFolder(repoId: string | number, collectionName: string, path: string, folderName: string) {
    return this.http.post<any>(`${this.apiUrl}/v1/collections/repo/${repoId}/${collectionName}/files/folder`, {
      path: path,
      folder: folderName,
    }, this.getAuthHeaders());
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
