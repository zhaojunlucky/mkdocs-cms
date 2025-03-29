import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import {ActivatedRoute, Router, RouterLink, RouterOutlet} from '@angular/router';
import { MatTabsModule } from '@angular/material/tabs';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { RepositoryService } from '../../services/repository.service';
import { Repository, Collection } from '../../services/repository.service';
import {RouteParameterService} from '../../services/routeparameter.service';
import {MatChipsModule} from '@angular/material/chips';

@Component({
  selector: 'app-repository-detail',
  standalone: true,
  imports: [
    CommonModule,
    MatTabsModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    RouterLink,
    RouterOutlet,
    MatChipsModule
  ],
  templateUrl: './repository-detail.component.html',
  styleUrls: ['./repository-detail.component.scss']
})
export class RepositoryDetailComponent implements OnInit {
  repository: Repository | null = null;
  collections: Collection[] = [];
  isLoading = true;
  error = '';
  selectedColName: string | null = '';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private repositoryService: RepositoryService,
    private routeParameterService: RouteParameterService
  ) { }

  ngOnInit(): void {
    this.route.paramMap.subscribe(params => {
      const id = params.get('id');
      if (id) {
        this.loadRepository(Number(id));
      } else {
        this.error = 'Repository ID is missing';
        this.isLoading = false;
      }
      const collectionName = params.get('collectionName');
      if (collectionName) {
        console.log('collectionName:', collectionName);
      }
    });
    this.routeParameterService.childId$.subscribe(childId => {
      this.selectedColName = childId;
    });
  }

  loadRepository(id: number): void {
    this.isLoading = true;
    this.error = '';

    this.repositoryService.getRepository(id).subscribe({
      next: (repo) => {
        this.repository = repo;
        this.loadCollections(id);
      },
      error: (err) => {
        console.error('Error loading repository:', err);
        this.error = 'Failed to load repository. Please try again later.';
        this.isLoading = false;
      }
    });
  }

  loadCollections(repositoryId: number): void {
    this.repositoryService.getRepositoryCollections(repositoryId).subscribe({
      next: (collections) => {
        this.collections = collections;
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Error loading collections:', err);
        this.error = 'Failed to load collections. Please try again later.';
        this.isLoading = false;
      }
    });
  }

  selectCollection(collection: Collection) {
    this.router.navigate(['/repositories', this.repository?.id, 'collection', collection.name]);
  }
}
