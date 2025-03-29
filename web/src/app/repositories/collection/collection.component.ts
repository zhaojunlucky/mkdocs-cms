import { Component, OnInit, Pipe, PipeTransform } from '@angular/core';
import { CommonModule, NgIf, NgFor, DatePipe } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';
import { MatMenuModule } from '@angular/material/menu';
import { RepositoryService, Repository, Collection } from '../../services/repository.service';
import { CollectionService, FileInfo } from '../../services/collection.service';
import { FrontMatterEditorComponent } from '../../markdown/front-matter-editor/front-matter-editor.component';
import { NuMarkdownComponent } from '@ng-util/markdown';
import {Observable} from 'rxjs';

@Pipe({
  name: 'fileSize',
  standalone: true
})
export class FileSizePipe implements PipeTransform {
  transform(bytes: number): string {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
}

@Component({
  selector: 'app-collection',
  standalone: true,
  imports: [
    CommonModule,
    NgIf,
    NgFor,
    DatePipe,
    RouterLink,
    FormsModule,
    MatProgressSpinnerModule,
    MatButtonModule,
    MatMenuModule,
    FileSizePipe
  ],
  templateUrl: './collection.component.html',
  styleUrls: ['./collection.component.scss']
})
export class CollectionComponent implements OnInit {
  repository: Repository | null = null;
  collection: Collection | null = null;
  error = '';
  repositoryId: string = '';
  collectionName: string = '';
  currentPath: string = '';
  files: FileInfo[] = [];
  pathSegments: { name: string; path: string }[] = [];

  selectedFile: FileInfo | null = null;


  // Front matter and markdown content
  frontMatter: Record<string, any> = {};
  markdownContent = '';

  // Folder operations
  isCreatingFolder = false;
  newFolderName = '';
  folderError = '';

  // Rename operations
  isRenaming = false;
  newName = '';
  renameError = '';

  // Global loading state for spinner overlay
  isLoading = true;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private repositoryService: RepositoryService,
    private collectionService: CollectionService
  ) { }

  ngOnInit(): void {
    if (this.route.parent) {
      this.repositoryId = this.route.parent.snapshot.paramMap.get('id') || '';
    }

    this.route.paramMap.subscribe(params => {
      const collectionName = params.get('collectionName');

      if (this.repositoryId && collectionName) {
        this.collectionName = collectionName;
        this.route.queryParams.subscribe(params => {
          this.currentPath = params['path'] || '';
          this.loadFiles();
        });

      } else {
        this.error = 'Invalid repository ID or collection name';
        this.isLoading = false;
      }
    });
  }

  loadFiles(): void {
    this.isLoading = true;
    let fileInfoObservable: Observable<FileInfo[]>;
    if (this.currentPath === '') {
      fileInfoObservable = this.collectionService.getCollectionFiles(this.repositoryId.toString(), this.collectionName);
    } else {
      fileInfoObservable = this.collectionService.getCollectionFilesInPath(this.repositoryId.toString(), this.collectionName, this.currentPath)
    }
    fileInfoObservable.subscribe({
      next: (files) => {
        this.files = files;
        this.updatePathSegments();
        this.isLoading = false;
      },
      error: (error) => {
        console.error('Error loading files:', error);
        this.error = 'Failed to load files. Please try again later.';
        this.isLoading = false;
      }
    });

  }

  updatePathSegments(): void {
    this.pathSegments = [];

    if (!this.currentPath) {
      return;
    }

    const pathParts = this.currentPath.split('/').filter(part => part.length > 0);
    let currentPath = '';

    // Add each path part
    for (let i = 0; i < pathParts.length; i++) {
      currentPath += (currentPath ? '/' : '') + pathParts[i];
      this.pathSegments.push({
        name: pathParts[i],
        path: currentPath
      });
    }
  }

  navigateToFolder(folder: FileInfo): void {
    if (!folder.is_dir) return;

    const path = folder.path;
    this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName],{
      queryParams: { path }
    });
  }


  // Folder operations
  openCreateFolderDialog(): void {
    this.isCreatingFolder = true;
    this.newFolderName = '';
    this.folderError = '';
  }

  cancelCreateFolder(): void {
    this.isCreatingFolder = false;
    this.newFolderName = '';
    this.folderError = '';
  }

  createFolder(): void {
    if (!this.repository || !this.collection) return;
    if (!this.newFolderName.trim()) {
      this.folderError = 'Please enter a folder name';
      return;
    }

    this.isLoading = true;
    this.folderError = '';

    this.collectionService.createFolder(
      this.repositoryId.toString(),
      this.collection.name,
      this.currentPath,
      this.newFolderName.trim()
    ).subscribe({
      next: () => {
        this.isLoading = false;
        this.isCreatingFolder = false;
        this.loadFiles(); // Refresh file list
      },
      error: (error) => {
        console.error('Error creating folder:', error);
        this.folderError = 'Failed to create folder. Please try again later.';
        this.isLoading = false;
      }
    });
  }

  // Rename operations
  openRenameDialog(file: FileInfo): void {
    this.selectedFile = file;
    this.isRenaming = true;
    this.newName = file.name;
    this.renameError = '';
  }

  cancelRename(): void {
    this.isRenaming = false;
    this.selectedFile = null;
    this.newName = '';
    this.renameError = '';
  }

  renameFileOrFolder(): void {
    if (!this.repository || !this.collection || !this.selectedFile) return;
    if (!this.newName.trim()) {
      this.renameError = 'Please enter a name';
      return;
    }

    this.isLoading = true;
    this.renameError = '';

    // Get the directory path from the current file path
    const currentPath = this.selectedFile.path;
    const lastSlashIndex = currentPath.lastIndexOf('/');
    const dirPath = lastSlashIndex !== -1 ? currentPath.substring(0, lastSlashIndex) : '';

    // Build the new path
    const newPath = dirPath ? `${dirPath}/${this.newName.trim()}` : this.newName.trim();

    this.collectionService.renameFile(
      this.repositoryId.toString(),
      this.collection.name,
      this.selectedFile.path,
      newPath
    ).subscribe({
      next: () => {
        this.isLoading = false;
        this.isRenaming = false;
        this.loadFiles(); // Refresh file list
      },
      error: (error) => {
        console.error('Error renaming:', error);
        this.renameError = 'Failed to rename. Please try again later.';
        this.isLoading = false;
      }
    });
  }

  // Delete operations
  deleteFileOrFolder(file: FileInfo): void {
    if (!this.repository || !this.collection) return;

    const isFolder = file.is_dir;
    const confirmMessage = isFolder
      ? `Are you sure you want to delete the folder "${file.name}" and all its contents?`
      : `Are you sure you want to delete the file "${file.name}"?`;

    if (confirm(confirmMessage)) {
      this.isLoading = true;

      this.collectionService.deleteFile(
        this.repositoryId.toString(),
        this.collection.name,
        file.path
      ).subscribe({
        next: () => {
          this.isLoading = false;
          this.loadFiles(); // Refresh file list
        },
        error: (error) => {
          console.error('Error deleting:', error);
          this.error = `Failed to delete ${isFolder ? 'folder' : 'file'}. Please try again later.`;
          this.isLoading = false;
        }
      });
    }
  }

  // Navigate to create file page
  openCreateFileDialog(): void {
    const path = this.currentPath ? this.currentPath : '';
    this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName, 'create'], {
      queryParams: { path }
    });
  }

  selectFile(file: FileInfo) {
    this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName, 'edit'], {
      queryParams: {
        path: file.path
      }
    });

  }
}
