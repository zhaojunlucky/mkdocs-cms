import { Component, Input, OnInit, ViewChild, forwardRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import {SimplemdeComponent, SimplemdeModule, SimplemdeOptions} from 'ngx-simplemde';
import { ControlValueAccessor, FormsModule, NG_VALUE_ACCESSOR, ReactiveFormsModule } from '@angular/forms';
import * as yaml from 'js-yaml';

@Component({
  selector: 'app-markdown-editor',
  standalone: true,
  imports: [CommonModule, SimplemdeModule, FormsModule, ReactiveFormsModule],
  templateUrl: './markdown-editor.component.html',
  styleUrls: ['./markdown-editor.component.scss'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => MarkdownEditorComponent),
      multi: true
    }
  ]
})
export class MarkdownEditorComponent implements OnInit, ControlValueAccessor {
  @ViewChild('simplemde', { static: true }) private simplemde!: SimplemdeComponent;

  @Input() disabled = false;
  @Input() placeholder = 'Type your markdown here...';

  markdownContent = '';
  frontMatter: Record<string, any> = {};
  private onChange: any = () => {};
  private onTouched: any = () => {};
  options: SimplemdeOptions = {
    indentWithTabs: false,
    tabSize: 2,
    shortcuts: {
      drawTable: "Cmd-Alt-T"
    },
    showIcons: ["code", "table"],
    spellChecker: true,
    autosave: {
      enabled: true,
      delay: 5000,
      uniqueId: 'markdown-editor'
    },
    renderingConfig: {
      singleLineBreaks: false,
      codeSyntaxHighlighting: true,
    },
  };

  ngOnInit(): void {
    // Additional initialization if needed

  }

  writeValue(value: string): void {
    if (value) {
      this.markdownContent = value;
      this.parseFrontMatter();
    } else {
      this.markdownContent = '';
      this.frontMatter = {};
    }
  }

  registerOnChange(fn: any): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }

  onContentChange(): void {
    // The content is already updated via ngModel binding
    this.parseFrontMatter();
    this.onChange(this.markdownContent);
    this.onTouched();
  }

  private parseFrontMatter(): void {
    try {
      // Check if the content has front matter (between --- delimiters)
      const frontMatterRegex = /^---\s*\n([\s\S]*?)\n---\s*\n/;
      const match = this.markdownContent.match(frontMatterRegex);

      if (match && match[1]) {
        this.frontMatter = yaml.load(match[1]) as Record<string, any>;
      } else {
        this.frontMatter = {};
      }
    } catch (error) {
      console.error('Error parsing front matter:', error);
      this.frontMatter = {};
    }
  }

  // Method to update front matter and regenerate the markdown content
  updateFrontMatter(newFrontMatter: Record<string, any>): void {
    this.frontMatter = { ...newFrontMatter };

    // Generate front matter YAML
    const frontMatterYaml = yaml.dump(this.frontMatter);

    // Check if the content already has front matter
    const frontMatterRegex = /^---\s*\n([\s\S]*?)\n---\s*\n/;

    if (frontMatterRegex.test(this.markdownContent)) {
      // Replace existing front matter
      this.markdownContent = this.markdownContent.replace(
        frontMatterRegex,
        `---\n${frontMatterYaml}---\n`
      );
    } else {
      // Add front matter to the beginning of the content
      this.markdownContent = `---\n${frontMatterYaml}---\n\n${this.markdownContent}`;
    }

    // Notify of the change
    this.onChange(this.markdownContent);
  }

  // Get the content without front matter
  getContentWithoutFrontMatter(): string {
    const frontMatterRegex = /^---\s*\n([\s\S]*?)\n---\s*\n/;
    return this.markdownContent.replace(frontMatterRegex, '');
  }
}
