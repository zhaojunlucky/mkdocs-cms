import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { TaskStatusComponent } from './task-status/task-status.component';

@NgModule({
  declarations: [

  ],
  imports: [
    CommonModule,
    RouterModule,
    TaskStatusComponent
  ],
  exports: [
    TaskStatusComponent
  ]
})
export class ComponentsModule { }
