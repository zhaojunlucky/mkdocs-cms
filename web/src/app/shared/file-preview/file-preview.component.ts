import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MaterialModule } from '../material/material.module';
import * as hljs from 'highlight.js';
import * as marked from 'marked';
import * as yaml from 'js-yaml';

@Component({
  selector: 'app-file-preview',
  standalone: true,
  imports: [CommonModule, MaterialModule],
  templateUrl: './file-preview.component.html',
  styleUrls: ['./file-preview.component.scss']
})
export class FilePreviewComponent implements OnChanges {
  @Input() fileContent: string = '';
  @Input() fileName: string = '';
  @Input() fileExtension: string = '';
  
  renderedContent: string = '';
  frontMatter: Record<string, any> | null = null;
  markdownContent: string = '';
  
  constructor() {
    // Configure marked renderer
    marked.setOptions({
      renderer: new marked.Renderer(),
      highlight: (code, lang) => {
        const language = hljs.getLanguage(lang) ? lang : 'plaintext';
        return hljs.highlight(code, { language }).value;
      },
      langPrefix: 'hljs language-',
      pedantic: false,
      gfm: true,
      breaks: false,
      sanitize: false,
      smartypants: false,
      xhtml: false
    });
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['fileContent'] || changes['fileExtension']) {
      this.renderContent();
    }
  }

  private renderContent(): void {
    if (!this.fileContent) {
      this.renderedContent = '';
      return;
    }

    // Handle markdown files with front matter
    if (this.fileExtension === 'md' || this.fileExtension === 'markdown') {
      this.processFrontMatter();
      this.renderedContent = marked.parse(this.markdownContent);
    } 
    // Handle YAML files
    else if (this.fileExtension === 'yml' || this.fileExtension === 'yaml') {
      try {
        const yamlObj = yaml.load(this.fileContent);
        this.renderedContent = `<pre class="hljs"><code>${hljs.highlight(this.fileContent, { language: 'yaml' }).value}</code></pre>`;
      } catch (e) {
        this.renderedContent = `<pre class="hljs"><code>${hljs.highlight(this.fileContent, { language: 'yaml' }).value}</code></pre>`;
      }
    }
    // Handle other file types with syntax highlighting
    else {
      let language = this.getLanguageFromExtension();
      this.renderedContent = `<pre class="hljs"><code>${hljs.highlight(this.fileContent, { language }).value}</code></pre>`;
    }
  }

  private processFrontMatter(): void {
    const frontMatterRegex = /^---\s*\n([\s\S]*?)\n---\s*\n([\s\S]*)$/;
    const match = this.fileContent.match(frontMatterRegex);
    
    if (match) {
      try {
        this.frontMatter = yaml.load(match[1]) as Record<string, any>;
        this.markdownContent = match[2];
      } catch (e) {
        this.frontMatter = null;
        this.markdownContent = this.fileContent;
      }
    } else {
      this.frontMatter = null;
      this.markdownContent = this.fileContent;
    }
  }

  private getLanguageFromExtension(): string {
    const extensionMap: Record<string, string> = {
      'js': 'javascript',
      'ts': 'typescript',
      'html': 'html',
      'css': 'css',
      'scss': 'scss',
      'json': 'json',
      'py': 'python',
      'go': 'go',
      'md': 'markdown',
      'markdown': 'markdown',
      'yml': 'yaml',
      'yaml': 'yaml',
      'sh': 'bash',
      'bash': 'bash',
      'txt': 'plaintext'
    };

    return extensionMap[this.fileExtension] || 'plaintext';
  }
}
