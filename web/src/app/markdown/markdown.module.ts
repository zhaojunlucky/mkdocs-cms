import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SimplemdeModule } from 'ngx-simplemde';

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    SimplemdeModule.forRoot({
      options: {
        placeholder: 'Type your markdown here...',
        spellChecker: false,
        autosave: {
          enabled: true,
          delay: 5000,
          uniqueId: 'markdown-editor'
        },
        renderingConfig: {
          singleLineBreaks: false,
          codeSyntaxHighlighting: true,
        },
        toolbar: [
          'bold', 'italic', 'heading', '|',
          'quote', 'unordered-list', 'ordered-list', '|',
          'link', 'image', '|',
          'preview', 'side-by-side', 'fullscreen', '|',
          'guide'
        ]
      }
    })
  ],
  exports: [
    SimplemdeModule
  ]
})
export class MarkdownModule { }
