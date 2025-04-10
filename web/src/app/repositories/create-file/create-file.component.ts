import { Component, OnInit } from '@angular/core';
import { CommonModule, NgIf, NgFor } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';
import { RepositoryService, Collection } from '../../services/repository.service';
import { CollectionService } from '../../services/collection.service';
import { FrontMatterEditorComponent } from '../../markdown/front-matter-editor/front-matter-editor.component';
import { NuMarkdownComponent } from '@ng-util/markdown';
import * as jsYaml from 'js-yaml';
import { MatInputModule} from '@angular/material/input';
import {MatIcon} from '@angular/material/icon';
import {MatTooltip} from '@angular/material/tooltip';
import {StrUtils} from '../../shared/utils/str.utils';

@Component({
  selector: 'app-create-file',
  standalone: true,
  imports: [
    CommonModule,
    NgIf,
    NgFor,
    RouterLink,
    FormsModule,
    MatProgressSpinnerModule,
    MatButtonModule,
    FrontMatterEditorComponent,
    NuMarkdownComponent,
    MatInputModule,
    MatIcon,
    MatTooltip
  ],
  templateUrl: './create-file.component.html',
  styleUrls: ['./create-file.component.scss']
})
export class CreateFileComponent implements OnInit {
  collection: Collection | null = null;
  error = '';
  repositoryId: string = '';
  collectionName: string = '';
  currentPath: string = '';
  pathSegments: { name: string; path: string }[] = [];

  // File creation properties
  fileName: string = '';
  fileError: string = '';
  isCreating: boolean = false;

  // Front matter and markdown content
  frontMatter: Record<string, any> = {};
  markdownContent: string = '';

  // Global loading state for spinner overlay
  isLoading: boolean = true;
  editor: any = null;

  // Editor options
  editorOptions = {
    theme: 'vs-light',
    language: 'markdown',
    lang: 'en_US',
    mode: 'wysiwyg',
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
          this.fileName = new Date().toISOString().slice(0, 10) + '-';
          this.loadCollection();
        });
      } else {
        this.error = 'Invalid repository ID or collection name';
        this.isLoading = false;
      }
    });
  }

  loadCollection(): void {
    this.repositoryService.getRepositoryCollections(this.repositoryId).subscribe({
      next: (collections) => {
        const foundCollection = collections.entries.find(c => c.name === this.collectionName);
        if (foundCollection) {
          this.collection = foundCollection;
          let bodyField = this.collection.fields?.find(f=>f.name === 'body')
          this.markdownContent = bodyField?.default || '';
          this.updatePathSegments();
          this.isLoading = false;
        } else {
          this.error = 'Collection not found';
          this.isLoading = false;
        }
      },
      error: (err) => {
        console.error('Error loading collections:', err);
        this.error = `Failed to load collection. ${StrUtils.stringifyHTTPErr(err)}`;
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


  onFrontMatterChange(newFrontMatter: Record<string, any>): void {
    this.frontMatter = { ...newFrontMatter };
  }

  onMarkdownChange(content: string): void {
    this.markdownContent = content;
  }

  onEditorReady(editor: any): void {
    this.editor = editor
  }

  createFile(): void {
    if (!this.repositoryId || !this.collection) return;
    if (!this.fileName.trim()) {
      this.fileError = 'Please enter a file name';
      return;
    }

    // Add .md extension if not present
    let finalFileName = this.fileName.trim();
    if (!finalFileName.endsWith('.md')) {
      finalFileName += '.md';
    }

    this.isLoading = true;
    this.isCreating = true;
    this.fileError = '';

    // Build file path
    const filePath = this.currentPath
      ? `${this.currentPath}/${finalFileName}`
      : finalFileName;

    // Build YAML front matter
    const yamlFrontMatter = `---\n${jsYaml.dump(this.frontMatter)}---\n`;
    const fileContent = `${yamlFrontMatter}${this.markdownContent}`;

    this.collectionService.uploadFile(
      this.repositoryId.toString(),
      this.collection.name,
      filePath,
      fileContent
    ).subscribe({
      next: () => {
        this.isLoading = false;
        this.isCreating = false;
        // Navigate back to collection view
        this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName], {
          queryParams: { path: this.currentPath }
        });
      },
      error: (error: any) => {
        console.error('Error creating file:', error);
        this.fileError = `Failed to create file. ${StrUtils.stringifyHTTPErr(error)}`;
        this.isLoading = false;
        this.isCreating = false;
      }
    });
  }

  cancel(): void {
    // Navigate back to collection view
    this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName], {
      queryParams: { path: this.currentPath }
    });
  }
}
