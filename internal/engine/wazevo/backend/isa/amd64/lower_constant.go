package amd64

import (
	"github.com/tetratelabs/wazero/internal/engine/wazevo/backend/regalloc"
	"github.com/tetratelabs/wazero/internal/engine/wazevo/ssa"
)

// lowerConstant allocates a new VReg and inserts the instruction to load the constant value.
func (m *machine) lowerConstant(instr *ssa.Instruction) (vr regalloc.VReg) {
	val := instr.Return()
	valType := val.Type()

	vr = m.c.AllocateVReg(valType)
	m.InsertLoadConstant(instr, vr)
	return
}

// InsertLoadConstant implements backend.Machine.
func (m *machine) InsertLoadConstant(instr *ssa.Instruction, vr regalloc.VReg) {
	val := instr.Return()
	valType := val.Type()
	v := instr.ConstantVal()

	bits := valType.Bits()
	if bits < 64 { // Clear the redundant bits just in case it's unexpectedly sign-extended, etc.
		v = v & ((1 << valType.Bits()) - 1)
	}

	switch valType {
	case ssa.TypeF32, ssa.TypeF64:
		m.lowerFconst(vr, v, bits == 64)
	case ssa.TypeI32, ssa.TypeI64:
		m.lowerIconst(vr, v, bits == 64)
	default:
		panic("BUG")
	}
}

func (m *machine) lowerFconst(dst regalloc.VReg, c uint64, _64 bool) {
	if c == 0 {
		xor := m.allocateInstr().asXmmUnaryRmR(sseOpcodeXorpd, newOperandReg(dst), dst)
		m.insert(xor)
	} else {
		var tmpType ssa.Type
		if _64 {
			tmpType = ssa.TypeI64
		} else {
			tmpType = ssa.TypeI32
		}
		tmpInt := m.c.AllocateVReg(tmpType)
		loadToGP := m.allocateInstr().asImm(tmpInt, c, _64)
		m.insert(loadToGP)

		movToXmm := m.allocateInstr().asGprToXmm(sseOpcodeMovq, newOperandReg(tmpInt), dst, _64)
		m.insert(movToXmm)
	}
}

func (m *machine) lowerIconst(dst regalloc.VReg, c uint64, _64 bool) {
	i := m.allocateInstr()
	i.asImm(dst, c, _64)
	m.insert(i)
}
