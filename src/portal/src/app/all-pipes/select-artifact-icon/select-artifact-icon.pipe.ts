import { Pipe, PipeTransform } from '@angular/core';
import { artifactImages, artifactDefault } from '../../project/repository/artifact/artifact';

@Pipe({
  name: 'selectArtifactIcon'
})
export class SelectArtifactIconPipe implements PipeTransform {

  transform(value: string, ...args: any[]): any {

      if (artifactImages.some(image => image === value)) {
        return 'images/artifact-' + value.toLowerCase() + '.svg';
      } else {
        return artifactDefault;
      }
  }

}
