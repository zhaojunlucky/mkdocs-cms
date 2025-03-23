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
      }
    })
  ],
  exports: [
    SimplemdeModule
  ]
})
export class MarkdownModule { }
