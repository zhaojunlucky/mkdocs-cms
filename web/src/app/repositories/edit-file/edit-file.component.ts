import {Component, NgZone, OnInit} from '@angular/core';
import { CommonModule, NgIf, NgForOf } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { NuMarkdownComponent } from '@ng-util/markdown';
import {Collection, RepositoryService} from '../../services/repository.service';
import {CollectionService} from '../../services/collection.service';
import * as yaml from 'js-yaml';
import {FrontMatterEditorComponent} from '../../markdown/front-matter-editor/front-matter-editor.component';
import {MatIcon} from '@angular/material/icon';
import {MatTooltip} from '@angular/material/tooltip';
import {ArrayResponse} from '../../shared/core/response';
import {StrUtils} from '../../shared/utils/str.utils';
import {CanComponentDeactivate} from '../../shared/guard/can-deactivate-form.guard';
import {Observable, of} from 'rxjs';

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
    NuMarkdownComponent,
    MatIcon,
    MatTooltip
  ],
  templateUrl: './edit-file.component.html',
  styleUrls: ['./edit-file.component.scss']
})
export class EditFileComponent implements OnInit, CanComponentDeactivate {
  repositoryId: string = '';
  collectionName: string = '';
  filePath: string = '';
  fileName: string = '';

  collection: Collection | null | undefined = null;

  isLoading: boolean = true;
  error: string = '';
  fileError: string = '';

  // Editor state
  _markdownContent: string = '';
  frontMatter: Record<string, any> = {};
  savingFile: boolean = false;

  // Path navigation
  pathSegments: PathSegment[] = [];
  editorRendered = false;
  changed = false;

  // Editor options
  editorOptions = {
    theme: 'vs-light',
    language: 'markdown',
    lang: 'en_US',
    icon: 'material',
    mode: 'wysiwyg',
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
    },
    after: ()=> this.zone.run(()=> {
      this.editorRendered = true;
    }),
  };

  editor: any = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private collectionService: CollectionService,
    private repositoryService: RepositoryService,
    private zone: NgZone
  ) {}

  get markdownContent(): string {
    return this._markdownContent;
  }

  set markdownContent(value: string) {
    this._markdownContent = value;
    this.changed = true;
  }

  ngOnInit(): void {
    if (this.route.parent) {
      this.repositoryId = this.route.parent.snapshot.paramMap.get('id') || '';
    }
    this.route.paramMap.subscribe(params => {
      this.collectionName = params.get('collectionName') || '';

      this.route.queryParamMap.subscribe(queryParams => {
        this.filePath = queryParams.get('path') || '';
        if (this.repositoryId && this.collectionName && this.filePath) {
          this.fileName = this.filePath.split('/').pop() || '';
          this.setupPathSegments();
          this.loadData();
        } else {
          this.error = 'Invalid repository ID or collection name';
          this.isLoading = false;
        }

      });


    });
  }

  loadData(): void {
    this.isLoading = true;
    this.error = '';
    if (!this.repositoryId || !this.collectionName || !this.filePath) {
      this.isLoading = false;
      return;
    }
    this.loadCollection();
    this.loadFileContent();
  }

  loadCollection(): void {
    this.repositoryService.getRepositoryCollections(this.repositoryId).subscribe({
      next: (collections: ArrayResponse<Collection>) => {
        this.collection = collections.entries.find(c => c.name === this.collectionName);
        if (this.collection) {
          this.loadFileContent()
        } else {
          this.error = `Failed to load collection: ${this.collectionName}`;
          this.isLoading = false;
        }
      },
      error: (err: any) => {
        this.error = `Failed to load collection: ${err.message || 'Unknown error'}`;
        this.isLoading = false;
      }
    });
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
          this._markdownContent = content.substring(endOfFrontMatter + 3).trim();
        } catch (e) {
          console.error('Error parsing front matter:', e);
          this.frontMatter = {};
          this._markdownContent = content;
        }
      } else {
        this.frontMatter = {};
        this._markdownContent = content;
      }
    } else {
      this.frontMatter = {};
      this._markdownContent = content;
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
    this.changed = true;
  }

  onFrontMatterInit(frontMatter: Record<string, any>): void {
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
        this.changed = false;
        // Navigate back to the collection view
        this.navigateToCollection();

      },
      error: (err: any) => {
        this.savingFile = false;
        this.fileError = `Failed to save file: ${StrUtils.stringifyHTTPErr(err)}`;
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

    const pathWithoutFile = this.filePath.substring(0, this.filePath.lastIndexOf('/'));
    this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName], {
      queryParams: {
        path: pathWithoutFile
      }
    });
  }

  canDeactivate(): Observable<boolean> | Promise<boolean> | boolean {
    if (!this.changed) {
      return true;
    }
    const confirmation = window.confirm('You have unsaved changes. Do you really want to leave?');
    return of(confirmation); // Return Observable<boolean>
  }
}
