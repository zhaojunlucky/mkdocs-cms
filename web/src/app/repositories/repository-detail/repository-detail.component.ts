import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { RepositoryService, Repository, Collection } from '../../services/repository.service';
import { CollectionService, FileInfo } from '../../services/collection.service';

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

  // New properties for sidenav and file browsing
  selectedCollection: Collection | null = null;
  currentPath: string = '';
  files: FileInfo[] = [];
  breadcrumbs: { name: string; path: string }[] = [];
  loadingFiles = false;

  constructor(
    private route: ActivatedRoute,
    private repositoryService: RepositoryService,
    private collectionService: CollectionService
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

  // Select a collection and load its files
  selectCollection(collection: Collection): void {
    this.selectedCollection = collection;
    this.currentPath = '';
    this.loadCollectionFiles();
    this.updateBreadcrumbs();
  }

  // Load files for the selected collection
  loadCollectionFiles() {
    if (this.selectedCollection) {
      // @ts-ignore
      this.collectionService.getCollectionFiles(this.repository.id, this.selectedCollection.name)
        .subscribe({
          next: (files) => {
            this.files = files;
            this.currentPath = '';
          },
          error: (error) => {
            console.error('Error loading collection files', error);
          }
        });
    }
  }

  // Navigate to a specific path in the collection
  navigateToPath(path: string) {
    if (this.selectedCollection) {
      // @ts-ignore
      this.collectionService.getCollectionFilesInPath(this.repository.id, this.selectedCollection.name, path)
        .subscribe({
          next: (files) => {
            this.files = files;
            this.currentPath = path;
          },
          error: (error) => {
            console.error('Error loading files for path', error);
          }
        });
    }
  }

  // Navigate to a folder
  navigateToFolder(folder: FileInfo): void {
    this.navigateToPath(folder.path);
    this.updateBreadcrumbs();
  }

  // Navigate to a specific breadcrumb
  navigateToBreadcrumb(breadcrumb: { name: string; path: string }): void {
    this.navigateToPath(breadcrumb.path);
    this.updateBreadcrumbs();
  }

  // Update breadcrumbs based on current path
  updateBreadcrumbs(): void {
    this.breadcrumbs = [];

    if (!this.currentPath) {
      return;
    }

    const pathParts = this.currentPath.split('/').filter(part => part.length > 0);
    let currentPath = '';

    // Add root
    this.breadcrumbs.push({ name: 'Root', path: '' });

    // Add each path part
    for (let i = 0; i < pathParts.length; i++) {
      currentPath += (currentPath ? '/' : '') + pathParts[i];
      this.breadcrumbs.push({
        name: pathParts[i],
        path: currentPath
      });
    }
  }

  // Get file icon based on file type
  getFileIcon(file: FileInfo): string {
    if (file.is_dir) {
      return 'folder';
    }

    switch (file.extension?.toLowerCase()) {
      case 'md':
        return 'description';
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
        return 'image';
      case 'pdf':
        return 'picture_as_pdf';
      default:
        return 'insert_drive_file';
    }
  }
}
