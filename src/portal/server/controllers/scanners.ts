import { Request, Response } from 'express';
import { Scanner } from '../../ng-swagger-gen/models/scanner';

export function getScanner(req: Request, res: Response) {
    const scanners: Scanner[] = [
        {
            name: 'demo',
            vendor: 'test',
            version: '1.0',
        },
    ];
    res.json(scanners);
}
