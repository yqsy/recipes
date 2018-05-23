struct SolveRequest {
    1: required string problem;
}


struct SolveReply {
    1: required bool ok;
    2: required string Result;
}

service SudoSolver {
    SolveReply Solve(1:SolveRequest solveRequest)
}
