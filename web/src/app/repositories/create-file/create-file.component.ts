import {Component, HostListener, NgZone, OnInit} from '@angular/core';

import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';
import { RepositoryService, Collection } from '../../services/repository.service';
import {CollectionService, FileInfo} from '../../services/collection.service';
import { FrontMatterEditorComponent } from '../../markdown/front-matter-editor/front-matter-editor.component';
import { NuMarkdownComponent } from '@ng-util/markdown';
import { MatInputModule} from '@angular/material/input';
import {MatIcon} from '@angular/material/icon';
import {MatTooltip} from '@angular/material/tooltip';
import {MatIconButton} from '@angular/material/button';
import {StrUtils} from '../../shared/utils/str.utils';
import {CanComponentDeactivate} from '../../shared/guard/can-deactivate-form.guard';
import {Observable, of} from 'rxjs';
import * as yaml from 'js-yaml';
import {PageTitleService} from '../../services/page.title.service';
import {VditorUploadService} from '../../services/vditor.upload.service';
import {MatSnackBar} from '@angular/material/snack-bar';
import {HttpHeaders} from '@angular/common/http';
import {ArrayResponse} from '../../shared/core/response';

@Component({
  selector: 'app-create-file',
  standalone: true,
  imports: [
    RouterLink,
    FormsModule,
    MatProgressSpinnerModule,
    MatButtonModule,
    FrontMatterEditorComponent,
    NuMarkdownComponent,
    MatInputModule,
    MatIcon,
    MatIconButton,
    MatTooltip
],
  templateUrl: './create-file.component.html',
  styleUrls: ['./create-file.component.scss']
})
export class CreateFileComponent implements OnInit, CanComponentDeactivate {
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
  _markdownContent: string = '';

  // Global loading state for spinner overlay
  isLoading: boolean = true;
  editor: any = null;
  editorRendered = false;

  // Editor options
  editorOptions = {
    theme: 'vs-light',
    language: 'markdown',
    lang: 'en_US',
    mode: 'wysiwyg',
    icon: 'material',
    tab: '    ',
    counter: {
      enable: true,
    },
    cache: {
      enable: false,
    },
    toolbar: [
      'emoji',
      'headings',
      'bold',
      'italic',
      'strike',
      'link',
      '|',
      'list',
      'ordered-list',
      'check',
      'outdent',
      'indent',
      '|',
      'quote',
      'line',
      'code',
      'inline-code',
      'insert-before',
      'insert-after',
      '|',
      'upload',
      'record',
      'table',
      '|',
      'undo',
      'redo',
      '|',
      'edit-mode',

      {
        name: 'more',
        toolbar: [
          'code-theme',
          'content-theme',
          'export',
          'outline',
          'preview',
          'devtools',
          'info',
          'help',
        ],
      },
    ],
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
    height: this.calculateEditorHeight(),
    after: ()=> this.zone.run(()=> {
      this.editorRendered = true;
      console.log("editor rendered");
    }),
  };
  _changed = true;
  private files: FileInfo[] = [];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private repositoryService: RepositoryService,
    private collectionService: CollectionService,
    private zone: NgZone,
    private pageTitleService: PageTitleService,
    private vditorUploadService: VditorUploadService,
    private snackBar: MatSnackBar
  ) {
    this.editorOptions = {...this.editorOptions, ...this.vditorUploadService.getVditorOptions()};
  }

  @HostListener('window:resize', ['$event'])
  onWindowResize() {
    // console.log("window resized");
    this.updateEditorHeight();
  }

  @HostListener('document:keydown', ['$event'])
  handleKeyDown(event: KeyboardEvent) {
    if ((event.ctrlKey || event.metaKey) && event.key === 's') {
      event.preventDefault(); // Prevent browser's default Save action
      if (this.changed) {
        this.createFile();
      }
    }
  }

  ngOnInit(): void {
    this.pageTitleService.title = 'Create File';
    if (this.route.parent) {
      this.repositoryId = this.route.parent.snapshot.paramMap.get('id') || '';
    }
    this.route.paramMap.subscribe(params => {
      const collectionName = params.get('collectionName');

      if (this.repositoryId && collectionName) {
        this.collectionName = collectionName;
        this.route.queryParams.subscribe(params => {
          this.currentPath = params['path'] || '';
          this.pageTitleService.title = `Create File - ${this.collectionName} - ${this.currentPath}`
          this.loadCollection();
        });
      } else {
        this.error = 'Invalid repository ID or collection name';
        this.isLoading = false;
      }
    });
  }

  get changed(): boolean {
    return this._changed;
  }

  set changed(value: boolean) {
    this._changed = value;
    let title = this.pageTitleService.title;
    if (this._changed && !title.startsWith('*')) {
      title = '* ' + title;
    } else if (!this._changed && title.startsWith('*')) {
      title = title.substring(1).trimStart()
    }
    this.pageTitleService.title = title;
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
          if (foundCollection?.file_name_generator?.type === 'sequence') {
            // load files
            this.loadFiles();
          } else {
            this.fileName = this.generateFileName();
          }
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
    this.changed = true;
  }

  get markdownContent(): string {
    return this._markdownContent;
  }

  set markdownContent(value: string) {
    this._markdownContent = value;
    this.changed = true;
  }

  onFrontMatterInit(frontMatter: Record<string, any>): void {
    this.frontMatter = frontMatter;
  }

  onEditorReady(editor: any): void {
    this.editor = editor;
    // Update height after editor is ready
    this.updateEditorHeight();
  }

  calculateEditorHeight(): number {
    // Calculate available height: viewport height minus nav (64px), footer (~53px), and padding/margins (~120px)
    const navHeight = 64;
    const footerHeight = 53;
    const paddingMargins = 100;

    const availableHeight = window.innerHeight - navHeight - footerHeight - paddingMargins;

    // Ensure minimum height of 300px
    return Math.max(300, availableHeight);
  }

  updateEditorHeight(): void {
    const newHeight = this.calculateEditorHeight();
    this.editorOptions = {
      ...this.editorOptions,
      height: newHeight
    };
  }

  generateFileContent(): string {
    // Only include front matter if it's not empty
    if (Object.keys(this.frontMatter).length === 0) {
      return this.markdownContent;
    }

    const frontMatterYaml = yaml.dump(this.frontMatter);
    return `---\n${frontMatterYaml}---\n\n${this.markdownContent}`;
  }

  createFile(): void {
    if (!this.repositoryId || !this.collection) return;
    if (!this.fileName.trim()) {
      this.fileError = 'Please enter a file name';
      this.showErrorMessage('Please enter a file name');
      return;
    }

    // Add .md extension if not present
    let finalFileName = this.fileName.trim();
    if (!finalFileName.endsWith('.md')) {
      finalFileName += '.md';
    }

    this.isCreating = true;
    this.editor.disabled();
    this.fileError = '';

    // Build file path
    const filePath = this.currentPath
      ? `${this.currentPath}/${finalFileName}`
      : finalFileName;

    const fileContent = this.generateFileContent()
    let headers = new HttpHeaders();
    if (this.frontMatter) {
      headers = headers.set('X-File-Front-Matter', 'true');
      headers = headers.set('X-File-Front-Matter-Draft', this.frontMatter['draft'] ? 'true' : 'false');
    }

    this.collectionService.uploadFile(
      this.repositoryId.toString(),
      this.collection.name,
      filePath,
      fileContent,
      headers
    ).subscribe({
      next: () => {
        this.isCreating = false;
        this.editor.enable();
        this.changed = false; // Reset changed
        // Navigate back to collection view
        this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName, 'edit'], {
          queryParams: { path: filePath }
        });
      },
      error: (error: any) => {
        console.error('Error creating file:', error);
        const errorMessage = `Failed to create file. ${StrUtils.stringifyHTTPErr(error)}`;
        this.fileError = errorMessage;
        this.showErrorMessage(errorMessage);
        this.isCreating = false;
        this.editor.enable();
      }
    });
  }

  cancel(): void {
    // Navigate back to collection view
    this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName], {
      queryParams: { path: this.currentPath }
    });
  }

  canDeactivate(): Observable<boolean> | Promise<boolean> | boolean {
    if (!this.changed) {
      return true;
    }
    const confirmation = window.confirm('You have unsaved changes. Do you really want to leave?');
    return of(confirmation); // Return Observable<boolean>
  }

  private showErrorMessage(message: string): void {
    this.snackBar.open(message, 'Close', {
      duration: 8000,
      panelClass: ['error-snackbar'],
      verticalPosition: 'top'
    });
  }

  private generateFileName() {
    if (this.collection?.file_name_generator) {
      switch (this.collection?.file_name_generator.type) {
        case 'date':
          return new Date().toISOString().slice(0, 10) + '-';
        case 'sequence': {
          if (this.files.length <= 0 && this.collection?.file_name_generator.first) {
            return this.collection.file_name_generator.first;
          }
          const pattern = /^(\d+)(\D+)?\.md$/;
          let maxNumber = 1;
          for (let i = 0; i < this.files.length; i++) {
            const match = this.files[i].name.match(pattern);
            if (match) {
              maxNumber = Math.max(maxNumber, parseInt(match[1]) + 1);
            }
          }
          return `${maxNumber}-`;
        }
        default:
          break;
      }
    }
    return '';
  }

  private loadFiles() {
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
        this.isLoading = false;
        this.fileName = this.generateFileName();
      },
      error: (error) => {
        console.error('Error loading files:', error);
        this.error = `Failed to load files. ${StrUtils.stringifyHTTPErr(error)}`;
        this.isLoading = false;
      }
    });
  }
}
