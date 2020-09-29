import { TargetExistsValidatorDirective } from './target-exists-directive';

describe('TargetExistsValidatorDirective', () => {
  it('should create an instance', () => {
    const directive = new TargetExistsValidatorDirective(null, null);
    expect(directive).toBeTruthy();
  });
});
