import { Component, OnInit } from '@angular/core';
import { CommonModule, NgIf, NgForOf } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { NuMarkdownComponent } from '@ng-util/markdown';
import {Collection, Repository, RepositoryService} from '../../services/repository.service';
import {CollectionService, FileInfo} from '../../services/collection.service';
import * as yaml from 'js-yaml';
import {FrontMatterEditorComponent} from '../../markdown/front-matter-editor/front-matter-editor.component';

interface PathSegment {
  name: string;
  path: string;
}

@Component({
  selector: 'app-edit-file',
  standalone: true,
  imports: [
    CommonModule,
    NgIf,
    NgForOf,
    FormsModule,
    RouterLink,
    MatProgressSpinnerModule,
    FrontMatterEditorComponent,
    NuMarkdownComponent
  ],
  templateUrl: './edit-file.component.html',
  styleUrls: ['./edit-file.component.scss']
})
export class EditFileComponent implements OnInit {
  repositoryId: string = '';
  collectionName: string = '';
  filePath: string = '';

  repository: Repository | null = null;
  collection: Collection | null | undefined = null;
  selectedFile: FileInfo | null = null;

  isLoading: boolean = true;
  error: string = '';
  fileError: string = '';

  // Editor state
  markdownContent: string = '';
  frontMatter: Record<string, any> = {};
  savingFile: boolean = false;

  // Path navigation
  pathSegments: PathSegment[] = [];

  // Editor options
  editorOptions = {
    theme: 'vs-light',
    language: 'markdown',
    lang: 'en_US',
    icon: 'material',
    counter: {
      enable: true,
    },
    cache: {
      enable: false,
    },
    preview: {
      hljs: {
        lineNumber: true,
      },
      markdown: {
        toc: true
      },
      actions: [
        "desktop"
      ]
    }
  };

  editor: any = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private collectionService: CollectionService
  ) {}

  ngOnInit(): void {
    this.route.paramMap.subscribe(params => {
      this.repositoryId = params.get('id') || '';
      this.collectionName = params.get('collectionName') || '';

      // Get the file path from the URL
      const url = this.router.url;
      const editIndex = url.indexOf('/edit/');
      if (editIndex !== -1) {
        this.filePath = url.substring(editIndex + 6);// +6 to skip '/edit/'
        if (!this.filePath) {
          this.router.navigate(['/404']);
        }
      }

      this.loadData();
    });
  }

  loadData(): void {
    this.isLoading = true;
    this.error = '';
    if (!this.repositoryId || !this.collectionName || !this.filePath) {
      this.isLoading = false;
      return;
    }
    this.loadFileContent();
  }

  loadFileContent(): void {
    this.isLoading = true;

    this.collectionService.getFileContent(this.repositoryId, this.collectionName, this.filePath).subscribe({
      next: (fileContent) => {
        // Parse the file content to separate front matter and markdown
        this.parseFileContent(fileContent);

        this.isLoading = false;

      },
      error: (err: any) => {
        this.error = `Failed to load file content: ${err.message || 'Unknown error'}`;
        this.isLoading = false;
      }
    });
  }

  parseFileContent(content: string): void {
    // Check if the file has front matter (starts with ---)
    if (content.startsWith('---')) {
      const endOfFrontMatter = content.indexOf('---', 3);
      if (endOfFrontMatter !== -1) {
        const frontMatterText = content.substring(3, endOfFrontMatter).trim();
        try {
          this.frontMatter = yaml.load(frontMatterText) as Record<string, any>;
          this.markdownContent = content.substring(endOfFrontMatter + 3).trim();
        } catch (e) {
          console.error('Error parsing front matter:', e);
          this.frontMatter = {};
          this.markdownContent = content;
        }
      } else {
        this.frontMatter = {};
        this.markdownContent = content;
      }
    } else {
      this.frontMatter = {};
      this.markdownContent = content;
    }
  }

  setupPathSegments(): void {
    this.pathSegments = [];

    if (this.filePath) {
      const segments = this.filePath.split('/');
      let currentPath = '';

      for (let i = 0; i < segments.length - 1; i++) {
        if (segments[i]) {
          currentPath += (currentPath ? '/' : '') + segments[i];
          this.pathSegments.push({
            name: segments[i],
            path: currentPath
          });
        }
      }
    }
  }

  onFrontMatterChange(frontMatter: Record<string, any>): void {
    this.frontMatter = frontMatter;
  }

  onEditorReady(editor: any): void {
    this.editor = editor;
  }

  saveFile(): void {
    if (this.savingFile) return;

    this.fileError = '';
    this.savingFile = true;

    // Combine front matter and markdown content
    const content = this.generateFileContent();

    this.collectionService.updateFileContent(
      this.repositoryId,
      this.collectionName,
      this.filePath,
      content
    ).subscribe({
      next: () => {
        this.savingFile = false;
        // Navigate back to the collection view
        this.navigateToCollection();
      },
      error: (err: any) => {
        this.savingFile = false;
        this.fileError = `Failed to save file: ${err.message || 'Unknown error'}`;
      }
    });
  }


  generateFileContent(): string {
    // Only include front matter if it's not empty
    if (Object.keys(this.frontMatter).length === 0) {
      return this.markdownContent;
    }

    const frontMatterYaml = yaml.dump(this.frontMatter);
    return `---\n${frontMatterYaml}---\n\n${this.markdownContent}`;
  }

  cancelEditing(): void {
    this.navigateToCollection();
  }

  navigateToCollection(): void {
    // Navigate back to the collection view
    if (this.filePath) {
      const pathWithoutFile = this.filePath.substring(0, this.filePath.lastIndexOf('/'));
      if (pathWithoutFile) {
        this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName, pathWithoutFile]);
      } else {
        this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName]);
      }
    } else {
      this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName]);
    }
  }
}
