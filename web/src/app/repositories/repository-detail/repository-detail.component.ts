import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import {RepositoryService, Repository, Collection, CollectionFieldDefinition} from '../../services/repository.service';
import { CollectionService, FileInfo } from '../../services/collection.service';
import { FormsModule } from '@angular/forms';
import { NgIf, NgFor, NgClass, DatePipe, DecimalPipe } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import * as jsYaml from 'js-yaml';
import { MarkdownModule } from '../../markdown/markdown.module';
import { MarkdownEditorComponent } from '../../markdown/markdown-editor/markdown-editor.component';
import { FrontMatterEditorComponent } from '../../markdown/front-matter-editor/front-matter-editor.component';
import {MatMenu, MatMenuTrigger} from '@angular/material/menu';
import {MatIcon} from '@angular/material/icon';

@Component({
  selector: 'app-repository-detail',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    NgIf,
    NgFor,
    NgClass,
    DatePipe,
    DecimalPipe,
    RouterLink,
    RouterModule,
    MatButtonModule,
    MarkdownModule,
    MarkdownEditorComponent,
    FrontMatterEditorComponent,
    MatMenuTrigger,
    MatIcon,
    MatMenu
  ],
  templateUrl: './repository-detail.component.html',
  styleUrls: ['./repository-detail.component.scss']
})
export class RepositoryDetailComponent implements OnInit {
  repository: Repository | null = null;
  collections: Collection[] = [];
  loadingRepo = true;
  loadingCollections = false;
  error = '';

  // New properties for sidenav and file browsing
  selectedCollection: Collection | null = null;
  currentPath: string = '';
  files: FileInfo[] = [];
  pathSegments: { name: string; path: string }[] = [];
  loadingFiles = false;

  // File editing properties
  isEditingFile = false;
  isCreatingFile = false;
  selectedFile: FileInfo | null = null;
  fileContent = '';
  loadingFileContent = false;
  savingFile = false;
  fileError = '';
  newFileName = '';

  // Front matter and markdown content
  frontMatter: Record<string, any> = {};
  markdownContent = '';

  constructor(
    private route: ActivatedRoute,
    private repositoryService: RepositoryService,
    private collectionService: CollectionService
  ) {}

  ngOnInit(): void {
    this.route.paramMap.subscribe(params => {
      const repoId = params.get('id');
      if (repoId) {
        this.loadRepository(repoId);
      } else {
        this.error = 'Invalid repository ID';
        this.loadingRepo = false;
      }
    });
  }

  loadRepository(id: string): void {
    this.loadingRepo = true;
    this.error = '';

    this.repositoryService.getRepository(Number(id)).subscribe({
      next: (repo) => {
        this.repository = repo;
        this.loadCollections(id);
      },
      error: (err) => {
        console.error('Error loading repository:', err);
        this.error = 'Failed to load repository. Please try again later.';
        this.loadingRepo = false;
      }
    });
  }

  loadCollections(repoId: string): void {
    this.loadingCollections = true;
    this.repositoryService.getRepositoryCollections(Number(repoId)).subscribe({
      next: (collections) => {
        this.collections = collections;
        this.loadingRepo = false;
        this.loadingCollections = false;
      },
      error: (err) => {
        console.error('Error loading collections:', err);
        this.error = 'Failed to load collections. Please try again later.';
        this.loadingRepo = false;
        this.loadingCollections = false;
      }
    });
  }

  // Select a collection and load its files
  selectCollection(collection: Collection): void {
    this.selectedCollection = collection;
    this.currentPath = '';
    this.loadCollectionFiles();
    this.updatePathSegments();
  }

  // Load files for the selected collection
  loadCollectionFiles() {
    if (!this.repository || !this.selectedCollection) return;

    this.loadingFiles = true;

    this.collectionService.getCollectionFiles(this.repository.id, this.selectedCollection.name, this.currentPath)
      .subscribe({
        next: (files) => {
          this.files = files;
          this.loadingFiles = false;
        },
        error: (error) => {
          console.error('Error loading collection files', error);
          this.loadingFiles = false;
        }
      });
  }

  // Navigate to a specific path in the collection
  navigateToPath(path: string) {
    if (!this.repository || !this.selectedCollection) return;

    this.loadingFiles = true;

    this.collectionService.getCollectionFiles(this.repository.id, this.selectedCollection.name, path)
      .subscribe({
        next: (files) => {
          this.files = files;
          this.currentPath = path;
          this.updatePathSegments();
          this.loadingFiles = false;
        },
        error: (error) => {
          console.error('Error loading files for path', error);
          this.loadingFiles = false;
        }
      });
  }

  // Navigate to a folder
  navigateToFolder(folder: FileInfo): void {
    this.navigateToPath(folder.path);
  }

  // Navigate to a specific breadcrumb
  navigateToBreadcrumb(path: string): void {
    this.navigateToPath(path);
  }

  // Update path segments based on current path
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

  // Open a file for editing
  openFile(file: FileInfo): void {
    if (!this.repository || !this.selectedCollection) return;

    this.loadingFileContent = true;
    this.fileError = '';

    this.collectionService.getFileContent(this.repository.id, this.selectedCollection.name, file.path)
      .subscribe({
        next: (content) => {
          this.fileContent = content;
          this.selectedFile = file;
          this.isEditingFile = true;
          this.loadingFileContent = false;

          // Parse YAML front matter if it exists
          this.parseYamlFrontMatter(content);
        },
        error: (error) => {
          console.error('Error loading file content', error);
          this.fileError = 'Failed to load file content. Please try again later.';
          this.loadingFileContent = false;
        }
      });
  }

  // Parse YAML front matter from markdown content
  parseYamlFrontMatter(content: string): void {
    // Reset metadata
    this.frontMatter = {};
    this.markdownContent = content;

    // Check if content has YAML front matter
    const yamlMatch = content.match(/^---\s*\n([\s\S]*?)\n---\s*\n([\s\S]*)$/);

    if (yamlMatch) {
      try {
        const yamlContent = yamlMatch[1];
        this.markdownContent = yamlMatch[2] || '';

        const metadata = jsYaml.load(yamlContent) as Record<string, any>;

        // Set front matter
        if (metadata) {
          this.frontMatter = metadata;
        }
      } catch (error) {
        console.error('Error parsing YAML front matter', error);
      }
    }
  }

  // Handle front matter changes from the editor
  onFrontMatterChange(newFrontMatter: Record<string, any>): void {
    this.frontMatter = { ...newFrontMatter };
  }

  // Handle markdown content changes
  onMarkdownChange(content: string): void {
    this.markdownContent = content;
  }

  // Save file with updated content and metadata
  saveFile(): void {
    if (!this.repository || !this.selectedCollection || !this.selectedFile) return;

    this.savingFile = true;
    this.fileError = '';

    // Build YAML front matter
    const yamlFrontMatter = `---\n${jsYaml.dump(this.frontMatter)}---\n`;
    const updatedContent = `${yamlFrontMatter}${this.markdownContent}`;

    this.collectionService.updateFileContent(
      this.repository.id,
      this.selectedCollection.name,
      this.selectedFile.path,
      updatedContent
    ).subscribe({
      next: () => {
        this.savingFile = false;
        this.isEditingFile = false;
        this.selectedFile = null;
        this.loadCollectionFiles(); // Refresh the file list
      },
      error: (error) => {
        console.error('Error saving file', error);
        this.fileError = 'Failed to save file. Please try again later.';
        this.savingFile = false;
      }
    });
  }

  // Create a new file in the current path
  createNewFile() {
    if (!this.repository || !this.selectedCollection) return;

    // Initialize front matter with default values from fields
    this.frontMatter = {};
    if (this.selectedCollection.fields) {
      this.selectedCollection.fields.forEach(field => {
        this.frontMatter[field.name] = null
      });
    }

    // Reset markdown content
    this.markdownContent = '';

    // Set creation mode
    this.isCreatingFile = true;
    this.isEditingFile = true;
    this.selectedFile = null;
    this.fileError = '';
  }

  // Save the new file
  saveNewFile() {
    if (!this.repository || !this.selectedCollection) return;
    if (!this.newFileName.trim()) {
      this.fileError = 'Please enter a file name';
      return;
    }

    // Add .md extension if not present
    const fileName = this.newFileName.trim().endsWith('.md') ?
      this.newFileName.trim() : `${this.newFileName.trim()}.md`;

    // Combine front matter and content
    const yamlFrontMatter = jsYaml.dump(this.frontMatter);
    const fileContent = `---\n${yamlFrontMatter}---\n\n${this.markdownContent}`;

    this.savingFile = true;
    this.fileError = '';

    // Create the file
    this.collectionService.uploadFile(
      this.repository.id,
      this.selectedCollection.name,
      this.currentPath ? `${this.currentPath}/${fileName}` : fileName,
      fileContent
    ).subscribe({
      next: () => {
        // Reset creation state
        this.isCreatingFile = false;
        this.isEditingFile = false;
        this.frontMatter = {};
        this.markdownContent = '';
        this.newFileName = '';
        this.savingFile = false;
        // Refresh the file list
        this.loadCollectionFiles();
      },
      error: (error) => {
        console.error('Error creating file:', error);
        this.fileError = 'Failed to create file. Please try again later.';
        this.savingFile = false;
      }
    });
  }

  // Cancel editing and close the editor
  cancelEditing() {
    this.isEditingFile = false;
    this.isCreatingFile = false;
    this.selectedFile = null;
    this.fileContent = '';
    this.frontMatter = {};
    this.markdownContent = '';
    this.fileError = '';
  }

  renameFile(file: FileInfo) {
    if (!this.repository || !this.selectedCollection) return;

    const newName = prompt(`Enter new name for ${file.name}:`, file.name);
    if (!newName || newName === file.name) return;

    // Construct the new path by replacing just the filename part
    const oldPath = file.path;
    const pathParts = oldPath.split('/');
    pathParts[pathParts.length - 1] = newName;
    const newPath = pathParts.join('/');

    this.collectionService.renameFile(this.repository.id, this.selectedCollection.name, oldPath, newPath)
      .subscribe({
        next: () => {
          // Reload the files to reflect the change
          this.loadCollectionFiles();
        },
        error: (error) => {
          console.error('Error renaming file:', error);
          alert(`Failed to rename file: ${error.error?.error || 'Unknown error'}`);
        }
      });
  }

  deleteFile(file: FileInfo) {
    if (!this.repository || !this.selectedCollection) return;

    if (confirm(`Are you sure you want to delete ${file.name}?`)) {
      this.collectionService.deleteFile(this.repository.id, this.selectedCollection.name, file.path)
        .subscribe({
          next: () => {
            // Refresh the file list after successful deletion
            this.loadCollectionFiles();
          },
          error: (error) => {
            console.error('Error deleting file:', error);
            this.error = 'Failed to delete file. Please try again later.';
          }
        });
    }
  }

  createNewFolder() {
    if (!this.repository || !this.selectedCollection) return;

    const folderName = prompt('Enter folder name:');
    if (!folderName) return;

    this.collectionService.createFolder(this.repository.id, this.selectedCollection.name, this.currentPath, folderName)
      .subscribe({
        next: () => {
          // Refresh the file list after successful folder creation
          this.loadCollectionFiles();
        },
        error: (error) => {
          console.error('Error creating folder:', error);
          this.error = 'Failed to create folder. ' + error.error?.error || 'Please try again later.';
        }
      });
  }
}
