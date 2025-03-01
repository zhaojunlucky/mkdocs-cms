import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { RepositoryService, Repository, Collection } from '../../services/repository.service';

@Component({
  selector: 'app-repository-detail',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './repository-detail.component.html',
  styleUrls: ['./repository-detail.component.scss']
})
export class RepositoryDetailComponent implements OnInit {
  repository: Repository | null = null;
  collections: Collection[] = [];
  loading = true;
  error = '';
  
  constructor(
    private route: ActivatedRoute,
    private repositoryService: RepositoryService
  ) {}
  
  ngOnInit(): void {
    this.route.paramMap.subscribe(params => {
      const repoId = Number(params.get('id'));
      if (repoId) {
        this.loadRepository(repoId);
      } else {
        this.error = 'Invalid repository ID';
        this.loading = false;
      }
    });
  }
  
  loadRepository(id: number): void {
    this.loading = true;
    this.error = '';
    
    this.repositoryService.getRepository(id).subscribe({
      next: (repo) => {
        this.repository = repo;
        this.loadCollections(id);
      },
      error: (err) => {
        console.error('Error loading repository:', err);
        this.error = 'Failed to load repository. Please try again later.';
        this.loading = false;
      }
    });
  }
  
  loadCollections(repoId: number): void {
    this.repositoryService.getRepositoryCollections(repoId).subscribe({
      next: (collections) => {
        this.collections = collections;
        this.loading = false;
      },
      error: (err) => {
        console.error('Error loading collections:', err);
        this.error = 'Failed to load collections. Please try again later.';
        this.loading = false;
      }
    });
  }
}
