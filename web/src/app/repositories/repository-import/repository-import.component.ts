import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { AuthService } from '../../auth/auth.service';
import { RepositoryService } from '../../services/repository.service';

interface GithubInstallation {
  id: number;
  account: {
    login: string;
    avatar_url: string;
  };
}

interface GithubRepository {
  id: number;
  name: string;
  full_name: string;
  html_url: string;
  description: string;
  private: boolean;
  owner: {
    login: string;
    avatar_url: string;
  };
  selected?: boolean; // For UI selection
}

@Component({
  selector: 'app-repository-import',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './repository-import.component.html',
  styleUrls: ['./repository-import.component.scss']
})
export class RepositoryImportComponent implements OnInit {
  installations: GithubInstallation[] = [];
  repositories: GithubRepository[] = [];
  selectedInstallation: number | null = null;
  loading = false;
  error = '';
  importSuccess = false;
  
  constructor(
    private http: HttpClient,
    private router: Router,
    private authService: AuthService,
    private repositoryService: RepositoryService
  ) {}
  
  ngOnInit(): void {
    this.loadInstallations();
  }
  
  loadInstallations(): void {
    this.loading = true;
    this.error = '';
    
    const headers = this.getAuthHeaders();
    this.http.get<GithubInstallation[]>('http://localhost:8080/api/v1/github/installations', { headers })
      .subscribe({
        next: (installations) => {
          this.installations = installations;
          this.loading = false;
          
          // Auto-select the first installation if there's only one
          if (this.installations.length === 1) {
            this.selectedInstallation = this.installations[0].id;
            this.loadRepositories(this.selectedInstallation);
          }
        },
        error: (err) => {
          console.error('Error loading installations:', err);
          this.error = 'Failed to load GitHub installations. Please make sure you have connected your GitHub account.';
          this.loading = false;
        }
      });
  }
  
  loadRepositories(installationId: number): void {
    this.loading = true;
    this.error = '';
    this.repositories = [];
    
    const headers = this.getAuthHeaders();
    this.http.get<GithubRepository[]>(`http://localhost:8080/api/v1/github/installations/${installationId}/repositories`, { headers })
      .subscribe({
        next: (repos) => {
          this.repositories = repos.map(repo => ({...repo, selected: false}));
          this.loading = false;
        },
        error: (err) => {
          console.error('Error loading repositories:', err);
          this.error = 'Failed to load repositories for this installation.';
          this.loading = false;
        }
      });
  }
  
  onInstallationChange(installationId: number): void {
    this.selectedInstallation = installationId;
    this.loadRepositories(installationId);
  }
  
  toggleSelectRepository(repo: GithubRepository): void {
    repo.selected = !repo.selected;
  }
  
  selectAll(select: boolean): void {
    this.repositories.forEach(repo => repo.selected = select);
  }
  
  hasSelectedRepositories(): boolean {
    return this.repositories.some(repo => repo.selected === true);
  }
  
  importRepositories(): void {
    if (!this.selectedInstallation) {
      this.error = 'Please select a GitHub installation.';
      return;
    }
    
    const selectedRepos = this.repositories.filter(repo => repo.selected);
    if (selectedRepos.length === 0) {
      this.error = 'Please select at least one repository to import.';
      return;
    }
    
    this.loading = true;
    this.error = '';
    
    const headers = this.getAuthHeaders();
    const payload = {
      repositories: selectedRepos.map(repo => repo.full_name)
    };
    
    this.http.post(`http://localhost:8080/api/v1/github/installations/${this.selectedInstallation}/import`, payload, { headers })
      .subscribe({
        next: () => {
          this.loading = false;
          this.importSuccess = true;
          setTimeout(() => {
            this.router.navigate(['/home']);
          }, 2000);
        },
        error: (err) => {
          console.error('Error importing repositories:', err);
          this.error = 'Failed to import repositories. Please try again later.';
          this.loading = false;
        }
      });
  }
  
  private getAuthHeaders(): HttpHeaders {
    const token = this.authService.currentToken;
    return new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
  }
}
