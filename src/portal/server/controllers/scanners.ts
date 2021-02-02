import { Scanner } from '../../src/app/base/left-side-nav/interrogation-services/scanner/scanner';
import { Request, Response } from "express";


export function getScanner(req: Request, res: Response) {
  const scanners: Scanner[] =  [new Scanner(), new Scanner()];
  res.json(scanners);
}


