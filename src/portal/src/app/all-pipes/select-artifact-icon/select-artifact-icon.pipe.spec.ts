import { SelectArtifactIconPipe } from './select-artifact-icon.pipe';

describe('SelectArtifactIconPipe', () => {
  let mockTypeImage = "IMAGE";
  let mockTypeChart = "CHART";
  let mockTypecnab = "CNAB";
  it('create an instance', () => {
    const pipe = new SelectArtifactIconPipe();
    expect(pipe).toBeTruthy();
  });
  it('it should success get adress of icon', () => {
    const pipe = new SelectArtifactIconPipe();
    expect(pipe.transform(mockTypeImage, '')).toBe('images/artifact-image.svg');
    expect(pipe.transform(mockTypeChart, '')).toBe('images/artifact-chart.svg');
    expect(pipe.transform(mockTypecnab, '')).toBe('images/artifact-cnab.svg');
    expect(pipe.transform("", '')).toBe('images/artifact-default.svg');

  });
});
