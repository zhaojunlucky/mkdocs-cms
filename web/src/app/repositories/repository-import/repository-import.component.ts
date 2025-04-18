import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../auth/auth.service';
import { catchError, finalize } from 'rxjs/operators';
import { of } from 'rxjs';
import {environment} from '../../../environments/environment';
import {ArrayResponse} from '../../shared/core/response';
import {StrUtils} from '../../shared/utils/str.utils';
import {PageTitleService} from '../../services/page.title.service';

interface GithubAccount {
  login: string;
  avatar_url: string;
}

interface GithubInstallation {
  id: number;
  account: GithubAccount;
}

interface GithubRepository {
  id: number;
  name: string;
  full_name: string;
  description: string;
  private: boolean;
  selected?: boolean;
}

interface GithubAppInfo {
  html_url: string;
  name: string;
}

@Component({
  selector: 'app-repository-import',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './repository-import.component.html',
  styleUrls: ['./repository-import.component.scss'],
  providers: [AuthService]
})
export class RepositoryImportComponent implements OnInit {
  loading = true;
  error = '';
  installations: GithubInstallation[] = [];
  repositories: GithubRepository[] = [];
  selectedInstallation: number | null = null;
  importSuccess = false;
  githubAppInfo: GithubAppInfo | null = null;

  private http = inject(HttpClient);
  private authService = inject(AuthService);
  private router = inject(Router);

  constructor(    private pageTitleService: PageTitleService
  ) {
  }

  ngOnInit(): void {
    this.pageTitleService.title = 'Import Repository';
    this.loadGithubAppInfo();
    this.loadInstallations();
  }

  loadGithubAppInfo(): void {
    this.http.get<GithubAppInfo>(`${environment.apiServer}/v1/github/app`)
      .pipe(
        catchError(err => {
          console.error('Error loading GitHub App info:', err);
          return of(null);
        })
      )
      .subscribe(appInfo => {
        if (appInfo) {
          this.githubAppInfo = appInfo;
        }
      });
  }

  navigateToGitHubAppInstallation(): void {
    window.open('https://github.com/apps/mkdocs-cms/', '_blank');
  }

  loadInstallations(): void {
    this.error = '';

    this.http.get<ArrayResponse<GithubInstallation>>(`${environment.apiServer}/v1/github/installations`)
      .subscribe({
        next: (installations: ArrayResponse<GithubInstallation>) => {
          this.installations = installations.entries;
          this.loading = false;

          // If no installations found, user needs to install the GitHub App
          if (this.installations.length === 0) {
            this.error = 'You need to install the GitHub App to your account first.';
          }
          // Auto-select the first installation if there's only one
          else if (this.installations.length === 1) {
            this.selectedInstallation = this.installations[0].id;
            this.loadRepositories(this.selectedInstallation);
          }
        },
        error: (err) => {
          console.error('Error loading installations:', err);
          this.error = `Failed to load GitHub installations. Please make sure you have connected your GitHub account. ${StrUtils.stringifyHTTPErr(err)}`;
          this.loading = false;
        }
      });
  }

  onInstallationChange(installationId: number): void {
    this.selectedInstallation = installationId;
    this.repositories = [];
    this.loadRepositories(installationId);
  }

  loadRepositories(installationId: number): void {
    this.loading = true;
    this.error = '';

    this.http.get<ArrayResponse<GithubRepository>>(
      `${environment.apiServer}/v1/github/installations/${installationId}/repositories`).subscribe({
      next: (repositories) => {
        this.repositories = repositories.entries.map(repo => ({
          ...repo,
          selected: false
        }));
        this.loading = false;

        if (this.repositories.length === 0) {
          this.error = 'No repositories found for this installation or all repositories have already been imported.';
        }
      },
      error: (err) => {
        console.error('Error loading repositories:', err);
        this.error = `Failed to load repositories. ${StrUtils.stringifyHTTPErr(err)}`;
        this.loading = false;
      }
    });
  }

  toggleSelectRepository(repo: GithubRepository): void {
    repo.selected = !repo.selected;
  }

  selectAll(selected: boolean): void {
    this.repositories.forEach(repo => repo.selected = selected);
  }

  hasSelectedRepositories(): boolean {
    return this.repositories.some(repo => repo.selected);
  }

  importRepositories(): void {
    if (!this.selectedInstallation) {
      this.error = 'No installation selected.';
      return;
    }

    const selectedRepos = this.repositories
      .filter(repo => repo.selected)
      .map(repo => repo.id);

    if (selectedRepos.length === 0) {
      this.error = 'No repositories selected.';
      return;
    }

    this.loading = true;
    this.error = '';

    this.http.post(
      `${environment.apiServer}/v1/github/installations/${this.selectedInstallation}/import`,
      { repositories: selectedRepos }
    ).pipe(
      finalize(() => this.loading = false)
    ).subscribe({
      next: () => {
        this.importSuccess = true;
        // Redirect to home page after 2 seconds
        setTimeout(() => {
          this.router.navigate(['/home']);
        }, 2000);
      },
      error: (err) => {
        console.error('Error importing repositories:', err);
        this.error = `Failed to import repositories. ${StrUtils.stringifyHTTPErr(err)}`;
      }
    });
  }
}
