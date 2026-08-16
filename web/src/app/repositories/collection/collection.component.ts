import { Component, OnInit, Pipe, PipeTransform, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';
import { MatMenuModule } from '@angular/material/menu';
import { RepositoryService, Repository, Collection } from '../../services/repository.service';
import { CollectionService, FileInfo } from '../../services/collection.service';
import {Observable, of} from 'rxjs';
import {catchError, map} from 'rxjs/operators';
import {MatIconModule} from '@angular/material/icon';
import {MatTooltip} from '@angular/material/tooltip';
import {MatCardModule} from '@angular/material/card';
import {MatFormField, MatInput, MatInputModule} from '@angular/material/input';
import {MatChipsModule} from '@angular/material/chips';
import {MatDialog, MatDialogModule} from '@angular/material/dialog';
import {ArrayResponse} from '../../shared/core/response';
import {StrUtils} from '../../shared/utils/str.utils';
import {PageTitleService} from '../../services/page.title.service';
import {SearchParser} from '../../shared/core/search.parser';
import {ConfirmDialogComponent} from '../../shared/dialogs/confirm-dialog.component';
import {TextInputDialogComponent} from '../../shared/dialogs/text-input-dialog.component';
import {FileNameDialogComponent} from '../../shared/dialogs/file-name-dialog.component';
import {FileNameParts, FileNameUtils} from '../../shared/utils/file-name.utils';

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
    DatePipe,
    RouterLink,
    FormsModule,
    MatProgressSpinnerModule,
    MatButtonModule,
    MatMenuModule,
    FileSizePipe,
    MatIconModule,
    MatTooltip,
    MatCardModule,
    MatInputModule,
    MatFormField,
    MatChipsModule,
    MatDialogModule
  ],
  templateUrl: './collection.component.html',
  changeDetection: ChangeDetectionStrategy.Eager,
  styleUrls: ['./collection.component.scss']
})
export class CollectionComponent implements OnInit {
  error = '';
  repositoryId: string = '';
  collectionName: string = '';
  currentPath: string = '';
  files: FileInfo[] = [];
  filteredFiles: FileInfo[] = [];
  searchTerm: string = '';
  pathSegments: { name: string; path: string }[] = [];

  hoveredFile: FileInfo | null = null;

  // Front matter and markdown content
  frontMatter: Record<string, any> = {};
  markdownContent = '';

  // Global loading state for spinner overlay
  isLoading = true;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private repositoryService: RepositoryService,
    private collectionService: CollectionService,
    private pageTitleService: PageTitleService,
    private dialog: MatDialog

  ) { }

  ngOnInit(): void {
    this.pageTitleService.title = 'Collection';
    if (this.route.parent) {
      this.repositoryId = this.route.parent.snapshot.paramMap.get('id') || '';
    }

    this.route.paramMap.subscribe(params => {
      const collectionName = params.get('collectionName');

      if (this.repositoryId && collectionName) {
        this.collectionName = collectionName;
        this.pageTitleService.title = `Collection - ${this.collectionName}`
        this.route.queryParams.subscribe(params => {
          this.currentPath = params['path'] || '';
          this.pageTitleService.title = `Collection - ${this.collectionName}${this.currentPath? ' - ' + this.currentPath : ''}`

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
    let fileInfoObservable: Observable<ArrayResponse<FileInfo>>;
    if (this.currentPath === '') {
      fileInfoObservable = this.collectionService.getCollectionFiles(this.repositoryId.toString(), this.collectionName);
    } else {
      fileInfoObservable = this.collectionService.getCollectionFilesInPath(this.repositoryId.toString(), this.collectionName, this.currentPath)
    }
    fileInfoObservable.subscribe({
      next: (files) => {
        this.files = files.entries;
        this.filterFiles();
        this.updatePathSegments();
        this.isLoading = false;
      },
      error: (error) => {
        console.error('Error loading files:', error);
        this.error = `Failed to load files. ${StrUtils.stringifyHTTPErr(error)}`;
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
    this.dialog.open(TextInputDialogComponent, {
      data: {
        title: 'Create Folder',
        label: 'Folder name',
        placeholder: 'Folder Name',
        confirmLabel: 'Create'
      },
      width: '420px'
    }).afterClosed().subscribe((folderName: string | null | undefined) => {
      if (folderName) {
        this.createFolder(folderName);
      }
    });
  }

  createFolder(folderName: string): void {
    if (!this.repositoryId || !this.collectionName) return;

    this.isLoading = true;

    this.collectionService.createFolder(
      this.repositoryId.toString(),
      this.collectionName,
      this.currentPath,
      folderName
    ).subscribe({
      next: () => {
        this.isLoading = false;
        this.loadFiles(); // Refresh file list
      },
      error: (error) => {
        console.error('Error creating folder:', error);
        this.error = `Failed to create folder. ${StrUtils.stringifyHTTPErr(error)}`;
        this.isLoading = false;
      }
    });
  }

  // Rename operations
  openRenameDialog(file: FileInfo): void {
    const nameParts = FileNameUtils.splitFileName(file.name);

    const dialogRef = this.dialog.open(FileNameDialogComponent, {
      data: {
        title: 'Rename File',
        prefix: nameParts.prefix,
        suffix: nameParts.suffix,
        extension: nameParts.extension,
        confirmLabel: 'Rename'
      },
      width: '520px'
    });

    this.collectionService.getFileContent(this.repositoryId, this.collectionName, file.path).subscribe({
      next: (content) => {
        dialogRef.componentInstance.data.markdownContent = content;
      },
      error: (error) => {
        console.error('Error loading file content for rename:', error);
      }
    });

    dialogRef.afterClosed().subscribe((suffix: string | null | undefined) => {
      if (!suffix) return;

      const newFileName = FileNameUtils.buildFileName(nameParts.prefix, suffix, nameParts.extension);
      if (newFileName !== file.name) {
        this.renameFile(file, nameParts, suffix);
      }
    });
  }

  renameFile(file: FileInfo, nameParts: FileNameParts, suffix: string): void {
    if (!this.repositoryId) return;
    const newFileName = FileNameUtils.buildFileName(nameParts.prefix, suffix, nameParts.extension);
    if (this.files.some(existingFile => existingFile.path !== file.path && existingFile.name === newFileName)) {
      this.dialog.open(ConfirmDialogComponent, {
        data: {
          title: 'File already exists',
          message: `${newFileName} already exists in this folder.`,
          confirmLabel: 'OK',
          cancelLabel: null,
          icon: 'error'
        }
      });
      return;
    }

    this.isLoading = true;

    // Get the directory path from the current file path
    const currentPath = file.path;
    const lastSlashIndex = currentPath.lastIndexOf('/');
    const dirPath = lastSlashIndex !== -1 ? currentPath.substring(0, lastSlashIndex) : '';

    // Build the new path
    const newPath = dirPath ? `${dirPath}/${newFileName}` : newFileName;

    this.collectionService.renameFile(
      this.repositoryId.toString(),
      this.collectionName,
      file.path,
      newPath
    ).subscribe({
      next: () => {
        this.isLoading = false;
        this.loadFiles(); // Refresh file list
      },
      error: (error) => {
        console.error('Error renaming:', error);
        this.error = `Failed to rename. ${StrUtils.stringifyHTTPErr(error)}`;
        this.isLoading = false;
      }
    });
  }

  openRenameFolderDialog(folder: FileInfo): void {
    if (!folder.is_dir) return;

    const parentPath = this.getParentPath(folder.path);
    this.dialog.open(TextInputDialogComponent, {
      data: {
        title: 'Rename Folder',
        label: 'Folder name',
        value: folder.name,
        placeholder: 'Folder Name',
        confirmLabel: 'Continue',
        validator: (value: string) => {
          const newFolderName = value.trim();
          if (!this.isValidFolderName(newFolderName)) {
            return 'Folder name cannot contain slashes, path traversal, or reserved names.';
          }
          if (newFolderName === folder.name) {
            return 'Enter a different folder name.';
          }
          if (this.files.some(file => file.path !== folder.path && file.name === newFolderName)) {
            return `An item named "${newFolderName}" already exists in this location.`;
          }
          return null;
        },
        asyncValidator: (value: string) => {
          const newFolderName = value.trim();
          const newPath = parentPath ? `${parentPath}/${newFolderName}` : newFolderName;
          return this.collectionService.folderExists(this.repositoryId.toString(), this.collectionName, newPath).pipe(
            map(response => {
              if (!response.exists) return null;
              return response.isDir
                ? `A folder named "${newFolderName}" already exists in this location.`
                : `A file named "${newFolderName}" already exists in this location.`;
            }),
            catchError(() => of('Could not check folder availability.'))
          );
        }
      },
      width: '420px'
    }).afterClosed().subscribe((folderName: string | null | undefined) => {
      if (!folderName) return;

      const newFolderName = folderName.trim();
      if (!this.isValidFolderName(newFolderName)) {
        this.showMessage('Invalid folder name', 'Folder name cannot contain slashes, path traversal, or reserved names.', 'error');
        return;
      }
      if (newFolderName === folder.name) {
        return;
      }
      if (this.files.some(file => file.path !== folder.path && file.name === newFolderName)) {
        this.showMessage('Folder already exists', `${newFolderName} already exists in this folder.`, 'error');
        return;
      }

      const newPath = parentPath ? `${parentPath}/${newFolderName}` : newFolderName;
      this.confirmRenameFolder(folder.path, newPath, () => this.renameFolder(folder.path, newPath));
    });
  }

  renameFolder(oldPath: string, newPath: string): void {
    if (!this.repositoryId || !this.collectionName) return;

    this.isLoading = true;
    this.collectionService.renameFolder(
      this.repositoryId.toString(),
      this.collectionName,
      oldPath,
      newPath
    ).subscribe({
      next: () => {
        this.isLoading = false;
        this.loadFiles();
      },
      error: (error) => {
        console.error('Error renaming folder:', error);
        this.error = `Failed to rename folder. ${StrUtils.stringifyHTTPErr(error)}`;
        this.isLoading = false;
      }
    });
  }

  deleteFolder(folder: FileInfo): void {
    if (!this.repositoryId || !this.collectionName || !folder.is_dir) return;

    this.dialog.open(ConfirmDialogComponent, {
      data: {
        title: 'Delete folder?',
        message: `This deletes "${folder.name}" and all files inside it. Relative links from other markdown files may point to deleted content.`,
        confirmLabel: 'Delete Folder',
        destructive: true,
        icon: 'delete'
      }
    }).afterClosed().subscribe((confirmed: boolean) => {
      if (!confirmed) return;

      this.isLoading = true;
      this.collectionService.deleteFolder(
        this.repositoryId.toString(),
        this.collectionName,
        folder.path
      ).subscribe({
        next: () => {
          this.isLoading = false;
          this.loadFiles();
        },
        error: (error) => {
          console.error('Error deleting folder:', error);
          this.error = `Failed to delete folder. ${StrUtils.stringifyHTTPErr(error)}`;
          this.isLoading = false;
        }
      });
    });
  }

  // Delete operations
  deleteFile(file: FileInfo): void {
    if (!this.repositoryId || !this.collectionName) return;

    const isFolder = file.is_dir;
    if (isFolder) {
      this.dialog.open(ConfirmDialogComponent, {
        data: {
          title: 'Cannot delete folder',
          message: 'Folders cannot be deleted from this view.',
          confirmLabel: 'OK',
          cancelLabel: null,
          icon: 'info'
        }
      });
      return;
    }

    this.dialog.open(ConfirmDialogComponent, {
      data: {
        title: 'Delete file',
        message: `Are you sure you want to delete the file "${file.name}"?`,
        confirmLabel: 'Delete',
        destructive: true,
        icon: 'delete'
      }
    }).afterClosed().subscribe((confirmed: boolean) => {
      if (!confirmed) return;

      this.isLoading = true;

      this.collectionService.deleteFile(
        this.repositoryId.toString(),
        this.collectionName,
        file.path
      ).subscribe({
        next: () => {
          this.isLoading = false;
          this.loadFiles(); // Refresh file list
        },
        error: (error) => {
          console.error('Error deleting:', error);
          this.error = `Failed to delete ${isFolder ? 'folder' : 'file'}. ${StrUtils.stringifyHTTPErr(error)}}`;
          this.isLoading = false;
        }
      });
    });
  }

  private confirmRenameFolder(oldPath: string, newPath: string, next: () => void): void {
    this.dialog.open(ConfirmDialogComponent, {
      data: {
        title: 'Rename folder?',
        message: `Renaming "${oldPath}" to "${newPath}" may break relative links and image paths in markdown files inside or outside this folder.`,
        confirmLabel: 'Rename Folder',
        icon: 'warning'
      }
    }).afterClosed().subscribe((confirmed: boolean) => {
      if (confirmed) {
        next();
      }
    });
  }

  private getParentPath(path: string): string {
    const lastSlashIndex = path.lastIndexOf('/');
    return lastSlashIndex === -1 ? '' : path.substring(0, lastSlashIndex);
  }

  private isValidFolderName(folderName: string): boolean {
    return !!folderName
      && folderName !== '.'
      && folderName !== '..'
      && !folderName.includes('/')
      && !folderName.includes('\\')
      && !folderName.includes('..');
  }

  private showMessage(title: string, message: string, icon: string): void {
    this.dialog.open(ConfirmDialogComponent, {
      data: {
        title,
        message,
        confirmLabel: 'OK',
        cancelLabel: null,
        icon
      }
    });
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

  // File search functionality
  filterFiles(): void {
    if (!this.searchTerm || this.searchTerm.trim() === '') {
      this.filteredFiles = [...this.files];
      return;
    }

    // Parse search syntax using the SearchParser
    const parser = new SearchParser(this.searchTerm);

    // Apply filters
    this.filteredFiles = this.files.filter(file => {
      // Check if file name matches content search
      const nameMatches = !parser.content || file.name.toLowerCase().includes(parser.content);

      // Check if file matches draft filter (if present)
      let draftMatches = true;
      if (parser.hasFilter('draft')) {
        const isDraft = parser.getFilter('draft') as boolean;
        draftMatches = file.is_draft === isDraft;
      }

      return nameMatches && draftMatches;
    });
  }

  // Clear search field and reset filtered files
  clearSearch(): void {
    this.searchTerm = '';
    this.filteredFiles = [...this.files];
  }

  onFileHover(file: FileInfo) {
    this.hoveredFile = file;
  }

  onFileLeave() {
    this.hoveredFile = null;
  }
}
